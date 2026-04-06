package leakage

import (
	"context"
	"fmt"
	"strconv"
)

const defaultDays = 30

// Service contains leakage detection use cases.
type Service struct {
	repo Repository
}

// NewService creates a leakage detection service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetReport returns leakage signals for the selected period.
func (s *Service) GetReport(ctx context.Context, userID string, days int) (*Report, error) {
	if days == 0 {
		days = defaultDays
	}
	if days < 7 || days > 90 {
		return nil, ErrInvalidDays
	}

	totalOutgoingRaw, err := s.repo.GetTotalOutgoing(ctx, userID, days)
	if err != nil {
		return nil, err
	}
	categorySpends, err := s.repo.ListCategorySpends(ctx, userID, days)
	if err != nil {
		return nil, err
	}
	repeatedPatterns, err := s.repo.ListRepeatedPatterns(ctx, userID, days)
	if err != nil {
		return nil, err
	}

	totalOutgoing := parseFloat(totalOutgoingRaw)
	report := &Report{
		Days:             days,
		TotalOutgoing:    formatFloat(totalOutgoing),
		CategorySpends:   categorySpends,
		RepeatedPatterns: repeatedPatterns,
	}

	for _, item := range categorySpends {
		amount := parseFloat(item.TotalAmount)
		if totalOutgoing <= 0 {
			continue
		}
		share := amount / totalOutgoing
		if share >= 0.45 {
			report.Findings = append(report.Findings, Finding{
				Type:     "category_concentration",
				Severity: severityFromShare(share),
				Title:    fmt.Sprintf("%s spending dominates recent outgoing flow", item.CategoryName),
				Summary:  fmt.Sprintf("%s accounts for %.0f%% of outgoing transactions in the last %d days.", item.CategoryName, share*100, days),
				Amount:   formatFloat(amount),
				Share:    formatPercent(share),
			})
		}
	}

	for _, item := range repeatedPatterns {
		total := parseFloat(item.TotalAmount)
		avg := parseFloat(item.AverageAmount)
		if item.TransactionCount >= 4 || (item.TransactionCount >= 3 && avg <= 150000) {
			report.Findings = append(report.Findings, Finding{
				Type:        "repeated_small_spend",
				Severity:    repeatedSeverity(item.TransactionCount, avg),
				Title:       fmt.Sprintf("Repeated outgoing pattern detected: %s", item.Description),
				Summary:     fmt.Sprintf("%q appeared %d times in the last %d days with average amount %s.", item.Description, item.TransactionCount, days, formatFloat(avg)),
				Amount:      formatFloat(total),
				Occurrences: item.TransactionCount,
			})
		}
	}

	report.Recommendations = buildRecommendations(report.Findings, totalOutgoing)
	return report, nil
}

func severityFromShare(share float64) string {
	switch {
	case share >= 0.6:
		return "high"
	case share >= 0.5:
		return "medium"
	default:
		return "low"
	}
}

func repeatedSeverity(count int, avgAmount float64) string {
	switch {
	case count >= 6 || avgAmount >= 250000:
		return "high"
	case count >= 4:
		return "medium"
	default:
		return "low"
	}
}

func buildRecommendations(findings []Finding, totalOutgoing float64) []string {
	items := make([]string, 0, 3)
	for _, finding := range findings {
		switch finding.Type {
		case "category_concentration":
			items = append(items, "Review the most dominant outgoing category and decide whether part of it can be capped or delayed.")
		case "repeated_small_spend":
			items = append(items, "Group repeated small purchases into one weekly review so they do not slip by unnoticed.")
		}
	}
	if totalOutgoing == 0 {
		items = append(items, "No outgoing activity was detected in the selected period, so there is nothing meaningful to flag yet.")
	}
	if len(items) == 0 {
		items = append(items, "No strong leakage signal was detected in the selected period. Keep categorizing spending consistently.")
	}
	return uniqueStrings(items)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func parseFloat(value string) float64 {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return number
}

func formatFloat(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

func formatPercent(value float64) string {
	return fmt.Sprintf("%.2f", value*100)
}

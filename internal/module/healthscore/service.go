package healthscore

import (
	"context"
	"fmt"
	"strconv"
)

const recentMonths = 3.0

// Service contains financial health scoring use cases.
type Service struct {
	repo Repository
}

// NewService creates a financial health scoring service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetReport returns the derived financial health score for the authenticated user.
func (s *Service) GetReport(ctx context.Context, userID string) (*Report, error) {
	snapshot, err := s.repo.GetSnapshot(ctx, userID)
	if err != nil {
		return nil, err
	}

	assets := mustParseFloat(snapshot.TotalAssets)
	debts := mustParseFloat(snapshot.TotalDebts)
	incoming := mustParseFloat(snapshot.RecentIncoming)
	outgoing := mustParseFloat(snapshot.RecentOutgoing)
	avgOutgoing := outgoing / recentMonths

	liquidityMonths := 0.0
	if avgOutgoing > 0 {
		liquidityMonths = assets / avgOutgoing
	}

	debtToAssetRatio := 0.0
	switch {
	case assets > 0:
		debtToAssetRatio = debts / assets
	case debts > 0:
		debtToAssetRatio = debts
	}

	cashFlowCoverage := 0.0
	switch {
	case outgoing > 0:
		cashFlowCoverage = incoming / outgoing
	case incoming > 0:
		cashFlowCoverage = 999
	}

	liquidityComponent := scoreLiquidity(liquidityMonths, outgoing)
	debtComponent := scoreDebt(debtToAssetRatio, debts)
	cashFlowComponent := scoreCashFlow(cashFlowCoverage, incoming, outgoing)

	totalScore := liquidityComponent.Score + debtComponent.Score + cashFlowComponent.Score

	return &Report{
		Score:   totalScore,
		Status:  scoreStatus(totalScore),
		Summary: scoreSummary(totalScore),
		Inputs: Inputs{
			TotalAssets:            formatFloat(assets),
			TotalDebts:             formatFloat(debts),
			RecentIncoming:         formatFloat(incoming),
			RecentOutgoing:         formatFloat(outgoing),
			AverageMonthlyOutgoing: formatFloat(avgOutgoing),
		},
		Metrics: Metrics{
			LiquidityMonths:  formatFloat(liquidityMonths),
			DebtToAssetRatio: formatFloat(debtToAssetRatio),
			CashFlowCoverage: formatFloat(cashFlowCoverage),
		},
		Components: []Component{
			liquidityComponent,
			debtComponent,
			cashFlowComponent,
		},
		Recommendations: recommendations(liquidityMonths, debtToAssetRatio, cashFlowCoverage, outgoing),
	}, nil
}

func scoreLiquidity(liquidityMonths, outgoing float64) Component {
	component := Component{
		Name:     "liquidity",
		MaxScore: 35,
	}
	switch {
	case outgoing == 0:
		component.Score = 35
		component.Description = "Spending is still low, so current assets provide strong short-term safety."
	case liquidityMonths >= 6:
		component.Score = 35
		component.Description = "Assets can cover at least 6 months of average spending."
	case liquidityMonths >= 3:
		component.Score = 28
		component.Description = "Assets can cover at least 3 months of average spending."
	case liquidityMonths >= 1:
		component.Score = 18
		component.Description = "Assets can cover about 1 month of average spending."
	case liquidityMonths > 0:
		component.Score = 8
		component.Description = "Liquidity buffer exists, but it is still thin."
	default:
		component.Score = 0
		component.Description = "No liquidity buffer is available against recent spending."
	}
	return component
}

func scoreDebt(debtToAssetRatio, debts float64) Component {
	component := Component{
		Name:     "debt_load",
		MaxScore: 35,
	}
	switch {
	case debts == 0:
		component.Score = 35
		component.Description = "No active debt reduces financial pressure."
	case debtToAssetRatio <= 0.25:
		component.Score = 30
		component.Description = "Debt is low compared with current assets."
	case debtToAssetRatio <= 0.5:
		component.Score = 22
		component.Description = "Debt is manageable, but worth keeping in check."
	case debtToAssetRatio <= 1:
		component.Score = 12
		component.Description = "Debt is approaching the size of current assets."
	default:
		component.Score = 4
		component.Description = "Debt is heavy relative to current assets."
	}
	return component
}

func scoreCashFlow(cashFlowCoverage, incoming, outgoing float64) Component {
	component := Component{
		Name:     "cash_flow",
		MaxScore: 30,
	}
	switch {
	case outgoing == 0 && incoming == 0:
		component.Score = 15
		component.Description = "Not enough movement yet to judge cash-flow stability."
	case outgoing == 0 && incoming > 0:
		component.Score = 30
		component.Description = "Recent income came in without outgoing pressure."
	case cashFlowCoverage >= 1.2:
		component.Score = 30
		component.Description = "Recent income comfortably covers recent spending."
	case cashFlowCoverage >= 1:
		component.Score = 24
		component.Description = "Recent income is roughly keeping up with spending."
	case cashFlowCoverage >= 0.8:
		component.Score = 15
		component.Description = "Recent income is slightly below spending."
	case cashFlowCoverage > 0:
		component.Score = 8
		component.Description = "Recent income covers only a small part of spending."
	default:
		component.Score = 0
		component.Description = "No recent incoming cash is covering current spending."
	}
	return component
}

func scoreStatus(score int) string {
	switch {
	case score >= 80:
		return "strong"
	case score >= 60:
		return "stable"
	case score >= 40:
		return "watch"
	default:
		return "risk"
	}
}

func scoreSummary(score int) string {
	switch {
	case score >= 80:
		return "Your current financial position looks strong."
	case score >= 60:
		return "Your financial position is stable, with a few areas worth monitoring."
	case score >= 40:
		return "Your finances are still workable, but they need closer attention."
	default:
		return "Your finances are under pressure and need corrective action soon."
	}
}

func recommendations(liquidityMonths, debtToAssetRatio, cashFlowCoverage, outgoing float64) []string {
	items := make([]string, 0, 3)
	if outgoing > 0 && liquidityMonths < 3 {
		items = append(items, "Build a larger cash buffer so assets can cover at least 3 months of spending.")
	}
	if debtToAssetRatio > 0.5 {
		items = append(items, "Prioritize reducing active debt to lower pressure on your balance sheet.")
	}
	if outgoing > 0 && cashFlowCoverage < 1 {
		items = append(items, "Increase net cash flow by raising income or trimming recurring spending.")
	}
	if len(items) == 0 {
		items = append(items, "Keep your current balance between liquidity, debt, and cash flow consistent.")
	}
	return items
}

func mustParseFloat(value string) float64 {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return number
}

func formatFloat(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

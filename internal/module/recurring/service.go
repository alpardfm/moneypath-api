package recurring

import (
	"context"
	"strings"
	"time"
)

// Service contains recurring use cases.
type Service struct {
	repo Repository
}

// NewService creates a recurring service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a recurring rule.
func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (*Rule, error) {
	rule, err := buildRule(userID, input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

// List returns active recurring rules for the authenticated user.
func (s *Service) List(ctx context.Context, userID string, options ListOptions) (*ListResult, error) {
	if options.Page <= 0 {
		options.Page = 1
	}
	if options.PageSize <= 0 {
		options.PageSize = 20
	}
	options.Type = strings.TrimSpace(options.Type)
	if options.Type != "" && !isValidType(options.Type) {
		return nil, ErrRuleValidation
	}
	return s.repo.List(ctx, userID, options)
}

// GetByID returns one recurring rule by id.
func (s *Service) GetByID(ctx context.Context, userID, ruleID string) (*Rule, error) {
	return s.repo.GetByID(ctx, userID, ruleID)
}

// Update updates recurring rule metadata.
func (s *Service) Update(ctx context.Context, userID, ruleID string, input UpdateInput) (*Rule, error) {
	rule, err := buildRule(userID, CreateInput(input))
	if err != nil {
		return nil, err
	}
	rule.ID = ruleID
	return s.repo.Update(ctx, rule)
}

// Inactivate hides a recurring rule from active use.
func (s *Service) Inactivate(ctx context.Context, userID, ruleID string) error {
	return s.repo.Inactivate(ctx, userID, ruleID)
}

// RunDue generates due mutations from active recurring rules.
func (s *Service) RunDue(ctx context.Context, userID string, now time.Time) (*RunDueResult, error) {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	return s.repo.RunDue(ctx, userID, now)
}

func buildRule(userID string, input CreateInput) (*Rule, error) {
	rule := &Rule{
		UserID:       userID,
		WalletID:     strings.TrimSpace(input.WalletID),
		Type:         strings.TrimSpace(input.Type),
		Amount:       strings.TrimSpace(input.Amount),
		Description:  strings.TrimSpace(input.Description),
		IntervalUnit: strings.TrimSpace(input.IntervalUnit),
		IntervalStep: input.IntervalStep,
		StartAt:      input.StartAt,
		EndAt:        input.EndAt,
	}
	if input.CategoryID != nil {
		value := strings.TrimSpace(*input.CategoryID)
		if value != "" {
			rule.CategoryID = &value
		}
	}

	if rule.IntervalStep == 0 {
		rule.IntervalStep = 1
	}
	if rule.WalletID == "" ||
		rule.Amount == "" ||
		rule.Description == "" ||
		rule.StartAt.IsZero() ||
		!isValidType(rule.Type) ||
		!isValidIntervalUnit(rule.IntervalUnit) ||
		rule.IntervalStep < 1 {
		return nil, ErrRuleValidation
	}
	if rule.EndAt != nil && rule.EndAt.Before(rule.StartAt) {
		return nil, ErrRuleValidation
	}

	rule.NextRunAt = rule.StartAt
	return rule, nil
}

func isValidType(value string) bool {
	return value == "masuk" || value == "keluar"
}

func isValidIntervalUnit(value string) bool {
	return value == "daily" || value == "weekly" || value == "monthly"
}

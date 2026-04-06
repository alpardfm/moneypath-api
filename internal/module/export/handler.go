package export

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/alpardfm/moneypath-api/internal/http/middleware"
	"github.com/alpardfm/moneypath-api/internal/http/response"
)

// Handler serves export endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates an export handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ExportMutationsCSV returns mutation history in CSV format.
func (h *Handler) ExportMutationsCSV(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.AuthUserID(r.Context())
	if !ok {
		response.Unauthorized(w)
		return
	}

	filter, err := parseMutationFilter(r)
	if err != nil {
		response.ValidationError(w, err.Error())
		return
	}

	rows, err := h.service.ExportMutations(r.Context(), userID, filter)
	if err != nil {
		if err == ErrInvalidFilter {
			response.ValidationError(w, err.Error())
			return
		}
		response.InternalError(w)
		return
	}

	filename := fmt.Sprintf("mutations-%s.csv", time.Now().Format("20060102-150405"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	_ = writer.Write([]string{
		"id",
		"wallet_name",
		"category_name",
		"debt_name",
		"type",
		"amount",
		"description",
		"related_to_debt",
		"happened_at",
		"created_at",
	})

	for _, item := range rows {
		_ = writer.Write([]string{
			item.ID,
			item.WalletName,
			item.CategoryName,
			item.DebtName,
			item.Type,
			item.Amount,
			item.Description,
			fmt.Sprintf("%t", item.RelatedToDebt),
			item.HappenedAt.Format(time.RFC3339),
			item.CreatedAt.Format(time.RFC3339),
		})
	}
}

func parseMutationFilter(r *http.Request) (MutationFilter, error) {
	filter := MutationFilter{
		Type:       r.URL.Query().Get("type"),
		WalletID:   r.URL.Query().Get("wallet_id"),
		CategoryID: r.URL.Query().Get("category_id"),
		DebtID:     r.URL.Query().Get("debt_id"),
	}

	if from := r.URL.Query().Get("from"); from != "" {
		value, err := time.Parse("2006-01-02", from)
		if err != nil {
			return MutationFilter{}, ErrInvalidFilter
		}
		filter.From = &value
	}
	if to := r.URL.Query().Get("to"); to != "" {
		value, err := time.Parse("2006-01-02", to)
		if err != nil {
			return MutationFilter{}, ErrInvalidFilter
		}
		filter.To = &value
	}

	return filter, nil
}

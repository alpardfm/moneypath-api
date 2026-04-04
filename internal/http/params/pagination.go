package params

import (
	"net/http"
	"strconv"
)

// Pagination contains common list pagination parameters.
type Pagination struct {
	Page     int
	PageSize int
}

// ParsePagination parses page and page_size query params with sane defaults.
func ParsePagination(r *http.Request) Pagination {
	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	pageSize := parsePositiveInt(r.URL.Query().Get("page_size"), 20)
	if pageSize > 100 {
		pageSize = 100
	}
	return Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}

func parsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

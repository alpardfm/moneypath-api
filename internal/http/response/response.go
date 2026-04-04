package response

import (
	"encoding/json"
	"net/http"
)

// Envelope defines the default API response format.
type Envelope struct {
	Success bool      `json:"success"`
	Data    any       `json:"data,omitempty"`
	Meta    any       `json:"meta,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

// APIError defines the standard API error shape.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// PaginationMeta describes paginated list response metadata.
type PaginationMeta struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
}

// JSON writes a JSON response with a consistent content type.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

// Success writes the standard success response envelope.
func Success(w http.ResponseWriter, status int, data any, meta any) {
	JSON(w, status, Envelope{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// Error writes the standard error response envelope.
func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, Envelope{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// InvalidJSON writes the standard invalid json response.
func InvalidJSON(w http.ResponseWriter) {
	Error(w, http.StatusBadRequest, "invalid_json", "invalid request body")
}

// ValidationError writes the standard validation error response.
func ValidationError(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "validation_error", message)
}

// Unauthorized writes the standard unauthorized response.
func Unauthorized(w http.ResponseWriter) {
	Error(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
}

// InternalError writes the standard internal server error response.
func InternalError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
}

// NewPaginationMeta builds pagination metadata.
func NewPaginationMeta(page, pageSize, totalItems int) PaginationMeta {
	totalPages := 0
	if pageSize > 0 {
		totalPages = (totalItems + pageSize - 1) / pageSize
	}
	return PaginationMeta{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

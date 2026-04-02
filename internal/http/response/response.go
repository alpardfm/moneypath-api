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

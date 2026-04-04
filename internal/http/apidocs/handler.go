package apidocs

import (
	_ "embed"
	"net/http"
)

//go:embed openapi.json
var openapiSpec []byte

// OpenAPI serves the embedded OpenAPI specification.
func OpenAPI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openapiSpec)
}

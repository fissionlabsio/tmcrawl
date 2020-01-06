package server

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse defines an error response returned upon any request failure.
type ErrorResponse struct {
	Error error `json:"error"`
}

func writeErrorResponse(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	bz, _ := json.Marshal(ErrorResponse{Error: err})
	_, _ = w.Write(bz)
}

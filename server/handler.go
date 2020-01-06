package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

const (
	methodGET = "GET"
)

// RegisterRoutes registers all HTTP routes with the provided mux router.
func RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/api/nodes", getNodesHandler()).Methods(methodGET)
}

func getNodesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

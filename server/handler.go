package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fissionlabsio/tmcrawl/crawl"
	"github.com/fissionlabsio/tmcrawl/db"
	"github.com/gorilla/mux"
)

const (
	methodGET = "GET"
)

// RegisterRoutes registers all HTTP routes with the provided mux router.
func RegisterRoutes(db db.DB, r *mux.Router) {
	r.HandleFunc("/api/nodes", getNodesHandler(db)).Methods(methodGET)
}

// PaginatedNodesResp defines a paginated search result of nodes.
type PaginatedNodesResp struct {
	Total int          `json:"total" yaml:"total"`
	Page  int          `json:"page" yaml:"page"`
	Limit int          `json:"limit" yaml:"limit"`
	Nodes []crawl.Node `json:"nodes" yaml:"nodes"`
}

func getNodesHandler(db db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pageStr := r.FormValue("page")
		limitStr := r.FormValue("limit")

		page := 1
		limit := 0

		if pageStr != "" {
			x, _ := strconv.Atoi(pageStr)
			if x <= 0 {
				writeErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid page query: %s", pageStr))
				return
			}

			page = x
		}

		if limitStr != "" {
			x, _ := strconv.Atoi(limitStr)
			if x <= 0 {
				writeErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid limit query: %s", limitStr))
				return
			}

			limit = x
		}

		nodes := []crawl.Node{}
		total := 0

		var err error
		db.IteratePrefix(crawl.NodeKeyPrefix, func(_, v []byte) bool {
			node := new(crawl.Node)
			err := node.Unmarshal(v)
			if err != nil {
				return true
			}

			total += 1
			nodes = append(nodes, *node)

			return false
		})

		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, fmt.Errorf("failed to query nodes: %w", err))
			return
		}

		start, end := paginate(len(nodes), page, limit, len(nodes))
		if start < 0 || end < 0 {
			nodes = []crawl.Node{}
		} else {
			nodes = nodes[start:end]
		}

		resp := PaginatedNodesResp{
			Page:  page,
			Limit: limit,
			Total: total,
			Nodes: nodes,
		}

		bz, err := json.Marshal(resp)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, fmt.Errorf("failed to encode response: %w", err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(bz)
	}
}

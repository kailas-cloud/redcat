package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) getCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("query")
	limitStr := r.URL.Query().Get("limit")

	var limit int64
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			limit = int64(v)
		}
	}

	cats, err := s.catSvc.Search(ctx, query, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := struct {
		Items []any `json:"items"`
	}{
		Items: make([]any, 0, len(cats)),
	}

	for _, c := range cats {
		resp.Items = append(resp.Items, c)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

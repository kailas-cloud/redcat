package api

import (
	"encoding/json"
	"net/http"
	"redcat/internal/service/categories"
	"strconv"
)

type Server struct {
	svc *categories.CategoryService
}

func NewServer(svc *categories.CategoryService) *Server {
	return &Server{svc: svc}
}

func (s *Server) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/categories", s.handleCategories)
}

func (s *Server) handleCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("query")
	limitStr := r.URL.Query().Get("limit")

	var limit int64
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			limit = int64(v)
		}
	}

	cats, err := s.svc.Search(ctx, query, limit)
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

package api

import (
	"encoding/json"
	"net/http"
	"redcat/internal/model"
)

func (s *Server) Places(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var p model.Place
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if err := validatePlace(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.placeSvc.Add(r.Context(), p); err != nil {
		http.Error(w, "failed to save place", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := struct {
		Item model.Place `json:"item"`
	}{
		Item: p,
	}

	_ = json.NewEncoder(w).Encode(resp)
}

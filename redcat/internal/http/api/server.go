package api

import (
	"net/http"
	"redcat/internal/service/categories"
	"redcat/internal/service/places"
)

type Server struct {
	catSvc   *categories.Service
	placeSvc *places.Service
}

func New(cat *categories.Service, ps *places.Service) *Server {
	return &Server{
		catSvc:   cat,
		placeSvc: ps,
	}
}

func (s *Server) Routes(mux *http.ServeMux) {
	mux.HandleFunc("/categories", s.Categories)
	mux.HandleFunc("/places", s.Places)
}

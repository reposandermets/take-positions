package controllers

import (
	"log"
	"net/http"

	"github.com/gorilla/context"
	"github.com/reposandermets/take-positions/api/middlewares"
)

func SetContext(f http.HandlerFunc, ctx int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		context.Set(r, "value", ctx)
		f(w, r)
	}
}

func (s *Server) initializeRoutes() {
	s.Router.HandleFunc("/", s.Home).Methods("GET")

	s.Router.HandleFunc("/signal", (middlewares.SetMiddlewareJSON(s.UpsertPosition))).Methods("POST")
	s.Router.HandleFunc("/signal", (middlewares.SetMiddlewareJSON(s.GetPosition))).Methods("GET")
}

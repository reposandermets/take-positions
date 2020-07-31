package controllers

import (
	"net/http"

	"github.com/reposandermets/take-positions/api/responses"
)

func (server *Server) Home(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, "Ok")
}

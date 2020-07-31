package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	core "github.com/reposandermets/take-positions/internal"

	"github.com/reposandermets/take-positions/api/responses"
)

func (server *Server) UpsertPosition(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, "OK")
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		println(err.Error())
		return
	}

	payload := core.Payload{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		println(err.Error())
		return
	}

	if payload.Signal == "Buy" {
		core.Q.Enqueue(payload)
	}
}

func (server *Server) GetPosition(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, core.F.FetchAccountState())
}

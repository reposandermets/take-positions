package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/reposandermets/take-positions/internal/account"
	"github.com/reposandermets/take-positions/internal/queue"

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

	payload := account.Payload{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		println(err.Error())
		return
	}

	queue.Q.Enqueue(payload)
}

func (server *Server) GetPosition(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, account.F.FetchAccountState("ETHUSD"))
}

package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/reposandermets/take-positions/internal/account"
	"github.com/reposandermets/take-positions/internal/logger"
	"github.com/reposandermets/take-positions/internal/queue"

	"github.com/reposandermets/take-positions/api/responses"
)

func getSignalString(sig int) string {
	if sig == 1 {
		return "Buy"
	}

	if sig == -1 {
		return "Sell"
	}

	if sig == 2 {
		return "ExitBuy"
	}

	if sig == -2 {
		return "ExitSell"
	}

	return "UNKNOWN"
}

func (server *Server) UpsertPosition(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, "OK")
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		logger.SendSlackNotification("ERROR UpsertPosition ioutil.ReadAll: " + err.Error())
		return
	}

	payload := account.Payload{}
	err = json.Unmarshal(body, &payload)
	if err != nil {
		println(err.Error())
		logger.SendSlackNotification("ERROR UpsertPosition json.Unmarshal: " + err.Error())
		return
	}

	payload.Ticker = strings.Replace(payload.Ticker, "/", "", -1)
	// should move to controller from here
	if payload.Ticker == "XBT" {
		payload.Ticker = "XBTUSD"
	} else if payload.Ticker == "ETH" {
		payload.Ticker = "ETHUSD"
	}

	payload.Signal = getSignalString(payload.Sig)

	if payload.Type != "Active" { // TODO use secret here instead
		logger.SendSlackNotification("Signal type mismatch: " + payload.Type)
		return
	}

	queue.Q.Enqueue(payload)
}

func (server *Server) GetPosition(w http.ResponseWriter, r *http.Request) {
	responses.JSON(w, http.StatusOK, account.F.FetchAccountState("ETHUSD"))
}

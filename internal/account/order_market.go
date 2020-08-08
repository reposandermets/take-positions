package account

import (
	"net/http"

	"github.com/zmxv/bitmexgo"
)

func (f *Flow) OrderMarket(accountState AccountState, payload Payload) (order bitmexgo.Order, res *http.Response, err error) {
	var params bitmexgo.OrderNewOpts
	params.OrdType.Set("Market")
	params.Side.Set(payload.Signal)
	params.OrderQty.Set(accountState.PositionSize)
	order, res, err = f.apiClient.OrderApi.OrderNew(f.auth, payload.Ticker, &params)

	return order, res, err
}

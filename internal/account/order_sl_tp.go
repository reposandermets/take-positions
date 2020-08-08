package account

import (
	"fmt"
	"math"
	"net/http"

	"github.com/zmxv/bitmexgo"
)

func (f *Flow) OrderSlTp(accountState AccountState, sl float64, tp float64) (orders []bitmexgo.Order, res *http.Response, err error) {
	oppositeSide := "Sell"
	if accountState.Position.CurrentQty < 0 {
		oppositeSide = "Buy"
	}

	orderQtyAbs := math.Abs(float64(accountState.Position.CurrentQty))
	var slOrder string
	if sl > 0 {
		slOrder = "{\"ordType\":\"Stop\",\"stopPx\":" + fmt.Sprintf("%g", sl) +
			",\"orderQty\":" + fmt.Sprintf("%g", orderQtyAbs) + ",\"side\":\"" + oppositeSide +
			"\",\"execInst\":\"Close,LastPrice\",\"symbol\":\"" + accountState.Position.Symbol + "\"}"
	}

	var tpOrder string
	if tp > 0 {
		tpOrder = "{\"ordType\":\"Limit\",\"price\":" + fmt.Sprintf("%g", tp) +
			",\"orderQty\":" + fmt.Sprintf("%g", orderQtyAbs) + ",\"side\":\"" + oppositeSide +
			"\",\"execInst\":\"ParticipateDoNotInitiate,ReduceOnly\",\"symbol\":\"" + accountState.Position.Symbol + "\"}"
	}

	var bulkOrders string
	if slOrder != "" && tpOrder != "" { // ? {"orders":[]}

		bulkOrders = `[` + slOrder + `,` + tpOrder + `]`
	} else if slOrder != "" && tpOrder == "" {
		bulkOrders = `[` + slOrder + `]`
	} else if slOrder == "" && tpOrder != "" {
		bulkOrders = `[` + tpOrder + `]`
	}

	var params bitmexgo.OrderNewBulkOpts
	params.Orders.Set(bulkOrders)

	println("")
	println("")
	println("")
	println("")
	println(params.Orders.Value())
	println("")
	println("")
	println("")
	orders, res, err = f.apiClient.OrderApi.OrderNewBulk(f.auth, &params)
	return orders, res, err
}

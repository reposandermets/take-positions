package account

import (
	"fmt"
	"math"
	"time"

	"github.com/reposandermets/take-positions/internal/logger"
	"github.com/zmxv/bitmexgo"
)

func (f *Flow) OrderSlTp(accountState AccountState, sl float64, tp float64) error {
	oppositeSide := "Sell"
	if accountState.Position.CurrentQty < 0 {
		oppositeSide = "Buy"
	}

	fullOrderQtyAbs := math.Abs(float64(accountState.Position.CurrentQty))

	slOrder := "{\"ordType\":\"Stop\",\"stopPx\":" + fmt.Sprintf("%g", sl) +
		",\"orderQty\":" + fmt.Sprintf("%g", fullOrderQtyAbs) + ",\"side\":\"" + oppositeSide +
		"\",\"execInst\":\"Close,LastPrice\",\"symbol\":\"" + accountState.Position.Symbol + "\"}"

	tpOrder := "{\"ordType\":\"Limit\",\"price\":" + fmt.Sprintf("%g", tp) +
		",\"orderQty\":" + fmt.Sprintf("%g", (fullOrderQtyAbs/2)) + ",\"side\":\"" + oppositeSide +
		"\",\"execInst\":\"ParticipateDoNotInitiate,ReduceOnly\",\"symbol\":\"" + accountState.Position.Symbol + "\"}"

	bulkOrders := `[` + slOrder + `,` + tpOrder + `]`

	var params bitmexgo.OrderNewBulkOpts
	params.Orders.Set(bulkOrders)

	println("###################")
	println("")
	println("")
	println("")
	println(params.Orders.Value())
	println("")
	println("")
	println("###################")

	return Retry(5, 3*time.Second, func() error {
		_, res, err := f.apiClient.OrderApi.OrderNewBulk(f.auth, &params)

		if res == nil || err != nil {
			println("ERROR OrderSlTp ", err.Error())
			return fmt.Errorf("network error: %v", 1)
		}

		s := res.StatusCode
		switch {
		case s >= 500:
			logger.SendSlackNotification("OrderSlTp http >= 500")
			return fmt.Errorf("server error: %v", s)
		case s == 429:
			time.Sleep(10 * time.Second)
			logger.SendSlackNotification("OrderSlTp http 429")
			return fmt.Errorf("Margin req http 429: %v", s)
		case s >= 400:
			logger.SendSlackNotification("OrderSlTp 4xx")
			return stop{fmt.Errorf("client error: %v", s)}
		default:
			return nil
		}
	})
}

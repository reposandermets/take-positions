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
	tpQty := math.Floor(fullOrderQtyAbs / 3.3)

	logger.SendLoggerNotification("INFO TP1: " + fmt.Sprintf("%g", tpQty) + " contracts at " + fmt.Sprintf("%g", tp) +
		" SL: " + fmt.Sprintf("%g", fullOrderQtyAbs) + " contracts at " + fmt.Sprintf("%g", sl))

	var positionId string
	positionId = fmt.Sprintf("%g", accountState.Position.AvgEntryPrice)

	slOrder := `{"ordType":"Stop","stopPx":` + fmt.Sprintf("%g", sl) + `,"text":"` + positionId + `"` +
		`,"orderQty":` + fmt.Sprintf("%g", fullOrderQtyAbs) + `,"side":"` + oppositeSide +
		`","execInst":"Close,LastPrice","symbol":"` + accountState.Position.Symbol + `"}`

	tpOrder := `{"ordType":"Limit","price":` + fmt.Sprintf("%g", tp) + `,"text":"` + positionId + `"` +
		`,"orderQty":` + fmt.Sprintf("%g", tpQty) + `,"side":"` + oppositeSide +
		`","execInst":"ParticipateDoNotInitiate,ReduceOnly","symbol":"` + accountState.Position.Symbol + `"}`

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
			logger.SendLoggerNotification("OrderSlTp http >= 500")
			return fmt.Errorf("server error: %v", s)
		case s == 429:
			time.Sleep(10 * time.Second)
			logger.SendLoggerNotification("OrderSlTp http 429")
			return fmt.Errorf("Margin req http 429: %v", s)
		case s >= 400:
			logger.SendLoggerNotification("OrderSlTp 4xx")
			return stop{fmt.Errorf("client error: %v", s)}
		default:
			return nil
		}
	})
}

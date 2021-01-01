package account

import (
	"fmt"
	"time"

	"github.com/reposandermets/take-positions/internal/logger"
	"github.com/zmxv/bitmexgo"
)

func (f *Flow) OrderMarket(positionSize int, payload Payload) error {
	var params bitmexgo.OrderNewOpts
	params.OrdType.Set("Market")
	params.Side.Set(payload.Signal)
	params.OrderQty.Set(positionSize)

	return Retry(5, 3*time.Second, func() error {
		_, res, err := f.apiClient.OrderApi.OrderNew(f.auth, payload.Ticker, &params)

		if res == nil || err != nil {
			println("OrderMarket ", err.Error())
			return fmt.Errorf("network error: %v", 1)
		}

		s := res.StatusCode
		switch {
		case s >= 500:
			logger.SendLoggerNotification("Position Close http >= 500")
			return fmt.Errorf("server error: %v", s)
		case s == 429:
			time.Sleep(10 * time.Second)
			logger.SendLoggerNotification("Position Close http 429")
			return fmt.Errorf("Margin req http 429: %v", s)
		case s >= 400:
			logger.SendLoggerNotification("Position Close 4xx")
			return stop{fmt.Errorf("client error: %v", s)}
		default:
			return nil
		}
	})
}

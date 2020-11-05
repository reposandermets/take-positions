package account

import (
	"fmt"
	"time"

	"github.com/reposandermets/take-positions/internal/logger"
	"github.com/zmxv/bitmexgo"
)

func (f *Flow) CancelOrders(ticker string) error {
	var params bitmexgo.OrderCancelAllOpts
	params.Symbol.Set(ticker)

	return Retry(3, 3*time.Second, func() error {
		_, res, err := f.apiClient.OrderApi.OrderCancelAll(f.auth, &params)

		if res == nil || err != nil {
			println("ERROR CancelOrders ", err.Error())
			return fmt.Errorf("network error: %v", 1)
		}

		s := res.StatusCode
		switch {
		case s >= 500:
			logger.SendSlackNotification("CancelOrders http >= 500")
			return fmt.Errorf("server error: %v", s)
		case s == 429:
			time.Sleep(10 * time.Second)
			logger.SendSlackNotification("CancelOrders http 429")
			return fmt.Errorf("Margin req http 429: %v", s)
		case s >= 400:
			logger.SendSlackNotification("CancelOrders 4xx")
			return stop{fmt.Errorf("client error: %v", s)}
		default:
			return nil
		}
	})
}

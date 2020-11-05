package account

import (
	"fmt"
	"time"

	"github.com/reposandermets/take-positions/internal/logger"
	"github.com/zmxv/bitmexgo"
)

func (f *Flow) ClosePosition(ticker string) error {
	var params bitmexgo.OrderNewOpts
	params.OrdType.Set("Market")
	params.ExecInst.Set("Close")

	return Retry(3, 3*time.Second, func() error {
		_, res, err := f.apiClient.OrderApi.OrderNew(f.auth, ticker, &params)

		if res == nil || err != nil {
			println("ClosePosition ", err.Error())
			return fmt.Errorf("network error: %v", 1)
		}

		s := res.StatusCode
		switch {
		case s >= 500:
			logger.SendSlackNotification("Position Close http >= 500")
			return fmt.Errorf("server error: %v", s)
		case s == 429:
			time.Sleep(10 * time.Second)
			logger.SendSlackNotification("Position Close http 429")
			return fmt.Errorf("Margin req http 429: %v", s)
		case s >= 400:
			logger.SendSlackNotification("Position Close 4xx")
			return stop{fmt.Errorf("client error: %v", s)}
		default:
			return nil
		}
	})
}

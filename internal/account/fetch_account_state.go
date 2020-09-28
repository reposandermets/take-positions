package account

import (
	"fmt"
	"net/http"
	"time"

	"github.com/reposandermets/take-positions/internal/logger"
	"github.com/zmxv/bitmexgo"
)

type stop struct {
	error
}

func retry(attempts int, sleep time.Duration, f func() error) error {
	if err := f(); err != nil {
		if s, ok := err.(stop); ok {
			return s.error
		}

		if attempts--; attempts > 0 {
			jitter := time.Duration(attempts)
			sleep = sleep / jitter
			time.Sleep(sleep)
			return retry(attempts, 2*sleep, f)
		}
		return err
	}

	return nil
}

func (f *Flow) FetchAccountState(symbol string) (accountState AccountState) {
	cPosition := make(chan PositionState)
	cMargin := make(chan MarginState)
	cTradeBin := make(chan TradeBinState)
	cTradeBinEth := make(chan TradeBinEthState)

	go func() {
		var params bitmexgo.TradeGetBucketedOpts
		params.BinSize.Set("1m")
		params.Partial.Set(true)
		params.Symbol.Set("ETHUSD")
		params.Reverse.Set(true)
		var tradeBins []bitmexgo.TradeBin
		var res *http.Response
		var err error
		var tradeBin bitmexgo.TradeBin

		retry(3, 3*time.Second, func() error {
			tradeBins, res, err = f.apiClient.TradeApi.TradeGetBucketed(f.auth, &params)
			if len(tradeBins) > 0 {
				tradeBin = tradeBins[0]
			}
			if tradeBins[0].Close > 0 {
				return nil
			}

			s := res.StatusCode
			switch {
			case s >= 500:
				logger.SendSlackNotification("ETHUSD tradeBins http >= 500")
				return fmt.Errorf("server error: %v", s)
			case s == 429:
				time.Sleep(10 * time.Second)
				logger.SendSlackNotification("ETHUSD tradeBins http 429")
				return fmt.Errorf("Margin req http 429: %v", s)
			case s >= 400:
				logger.SendSlackNotification("ETHUSD tradeBins http >= 400")
				return stop{fmt.Errorf("client error: %v", s)}
			case tradeBins[0].Close == 0:
				logger.SendSlackNotification("ETHUSD tradeBins[0].Close is 0")
				return fmt.Errorf("ETHUSD tradeBins[0].Close is 0")
			default:
				return nil
			}
		})
		logger.SendSlackNotification("ETHUSD tradeBins[0].Close: " + fmt.Sprintf("%f", tradeBins[0].Close))
		cTradeBinEth <- TradeBinEthState{
			TradeBinEth: tradeBin,
			Error:       err,
		}
	}()

	go func() {
		var params bitmexgo.PositionGetOpts
		params.Filter.Set("{\"symbol\":\"" + symbol + "\"}")
		var positions []bitmexgo.Position
		var position bitmexgo.Position
		var res *http.Response
		var err error
		var side string
		var hasOpenPosition bool
		retry(3, 3*time.Second, func() error {
			positions, res, err = f.apiClient.PositionApi.PositionGet(f.auth, &params)
			var position bitmexgo.Position
			side = ""
			hasOpenPosition = false
			if len(positions) > 0 {
				position = positions[0]
				hasOpenPosition = position.IsOpen
				if position.CurrentQty > 0 {
					side = "Buy"
				} else if position.CurrentQty < 0 {
					side = "Sell"
				}
			}

			s := res.StatusCode
			switch {
			case s >= 500:
				logger.SendSlackNotification("PositionGet http >= 500")
				return fmt.Errorf("server error: %v", s)
			case s == 429:
				time.Sleep(10 * time.Second)
				logger.SendSlackNotification("PositionGet http 429")
				return fmt.Errorf("Margin req http 429: %v", s)
			case s >= 400:
				logger.SendSlackNotification("PositionGet http >= 400")
				return stop{fmt.Errorf("client error: %v", s)}
			default:
				return nil
			}
		})

		cPosition <- PositionState{
			HasOpenPosition: hasOpenPosition,
			Side:            side,
			Position:        position,
			Error:           err,
		}
	}()

	go func() {
		var margin bitmexgo.Margin
		var res *http.Response
		var err error
		retry(3, 3*time.Second, func() error {
			margin, res, err = f.apiClient.UserApi.UserGetMargin(f.auth, nil)
			s := res.StatusCode
			switch {
			case s >= 500:
				logger.SendSlackNotification("Margin req http >= 500")
				return fmt.Errorf("server error: %v", s)
			case s == 429:
				time.Sleep(10 * time.Second)
				logger.SendSlackNotification("Margin req http 429")
				return fmt.Errorf("Margin req http 429: %v", s)
			case s >= 400:
				logger.SendSlackNotification("Margin req http >= 400")
				return stop{fmt.Errorf("client error: %v", s)}
			default:
				return nil
			}
		})
		cMargin <- MarginState{
			Margin: margin,
			Error:  err,
		}
	}()

	go func() {
		var params bitmexgo.TradeGetBucketedOpts
		params.BinSize.Set("1m")
		params.Partial.Set(true)
		params.Symbol.Set("XBTUSD")
		params.Reverse.Set(true)
		var tradeBins []bitmexgo.TradeBin
		var res *http.Response
		var err error
		var tradeBin bitmexgo.TradeBin
		retry(3, 3*time.Second, func() error {
			tradeBins, res, err = f.apiClient.TradeApi.TradeGetBucketed(f.auth, &params)

			if len(tradeBins) > 0 {
				tradeBin = tradeBins[0]
			}
			if tradeBins[0].Close > 0 {
				return nil
			}

			s := res.StatusCode
			switch {
			case s >= 500:
				logger.SendSlackNotification("XBTUSD tradeBins http >= 500")
				return fmt.Errorf("server error: %v", s)
			case s == 429:
				time.Sleep(10 * time.Second)
				logger.SendSlackNotification("XBTUSD tradeBins http 429")
				return fmt.Errorf("Margin req http 429: %v", s)
			case s >= 400:
				logger.SendSlackNotification("XBTUSD tradeBins http >= 400")
				return stop{fmt.Errorf("client error: %v", s)}
			case tradeBins[0].Close == 0:
				logger.SendSlackNotification("XBTUSD tradeBins[0].Close is 0")
				return fmt.Errorf("XBTUSD tradeBins[0].Close is 0")
			default:
				return nil
			}
		})

		cTradeBin <- TradeBinState{
			TradeBin: tradeBin,
			Error:    err,
		}
	}()

	for i := 0; i < 4; i++ {
		select {

		case msgPosition := <-cPosition:
			accountState.Side = msgPosition.Side
			accountState.HasOpenPosition = msgPosition.HasOpenPosition
			accountState.Position = msgPosition.Position
			accountState.PositionError = msgPosition.Error

		case msgMargin := <-cMargin:
			accountState.Margin = msgMargin.Margin
			accountState.MarginError = msgMargin.Error

		case msgTradeBin := <-cTradeBin:
			accountState.TradeBin = msgTradeBin.TradeBin
			accountState.TradeBinError = msgTradeBin.Error

		case msgTradeBinEth := <-cTradeBinEth:
			accountState.TradeBinEth = msgTradeBinEth.TradeBinEth
			accountState.TradeBinEthError = msgTradeBinEth.Error
		}
	}

	return accountState
}

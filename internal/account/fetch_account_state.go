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

func Retry(attempts int, sleep time.Duration, f func() error) error {
	if err := f(); err != nil {
		if s, ok := err.(stop); ok {
			return s.error
		}

		if attempts--; attempts > 0 {
			jitter := time.Duration(attempts)
			sleep = sleep / jitter
			time.Sleep(sleep)
			return Retry(attempts, 2*sleep, f)
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
		// give BM time to calculate the Close price
		for time.Now().Second() >= 0 && time.Now().Second() < 15 {
			time.Sleep(time.Second)
		}
		var params bitmexgo.TradeGetBucketedOpts
		params.BinSize.Set("1m")
		params.Partial.Set(true)
		params.Symbol.Set("XBTUSD")
		params.Reverse.Set(true)
		var tradeBins []bitmexgo.TradeBin
		var res *http.Response
		var err error
		var tradeBin bitmexgo.TradeBin
		Retry(5, 3*time.Second, func() error {
			tradeBins, res, err = f.apiClient.TradeApi.TradeGetBucketed(f.auth, &params)
			if res == nil || err != nil {
				println("TradeGetBucketed XBT", err.Error())
				return fmt.Errorf("network error: %v", 1)
			}
			if len(tradeBins) > 0 {
				tradeBin = tradeBins[0]
			}
			if tradeBins[0].Close > 0 {
				return nil
			}

			s := res.StatusCode
			switch {
			case s >= 500:
				logger.SendLoggerNotification("XBTUSD tradeBins http >= 500")
				return fmt.Errorf("server error: %v", s)
			case s == 429:
				time.Sleep(10 * time.Second)
				logger.SendLoggerNotification("XBTUSD tradeBins http 429")
				return fmt.Errorf("Margin req http 429: %v", s)
			case s >= 400:
				logger.SendLoggerNotification("XBTUSD tradeBins http >= 400")
				return stop{fmt.Errorf("client error: %v", s)}
			case tradeBins[0].Close == 0:
				logger.SendLoggerNotification("XBTUSD tradeBins[0].Close is 0")
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

	go func() {
		// give BM time to calculate the Close price
		for time.Now().Second() >= 0 && time.Now().Second() < 16 {
			time.Sleep(time.Second)
		}
		var params bitmexgo.TradeGetBucketedOpts
		params.BinSize.Set("1m")
		params.Partial.Set(true)
		params.Symbol.Set("ETHUSD")
		params.Reverse.Set(true)
		var tradeBins []bitmexgo.TradeBin
		var res *http.Response
		var err error
		var tradeBin bitmexgo.TradeBin

		Retry(5, 3*time.Second, func() error {
			tradeBins, res, err = f.apiClient.TradeApi.TradeGetBucketed(f.auth, &params)
			if res == nil || err != nil {
				println("TradeGetBucketed ETH", err.Error())
				return fmt.Errorf("network error: %v", 1)
			}

			if len(tradeBins) > 0 {
				tradeBin = tradeBins[0]
			}
			if tradeBins[0].Close > 0 {
				return nil
			}

			s := res.StatusCode
			switch {
			case s >= 500:
				logger.SendLoggerNotification("ETHUSD tradeBins http >= 500")
				return fmt.Errorf("server error: %v", s)
			case s == 429:
				time.Sleep(10 * time.Second)
				logger.SendLoggerNotification("ETHUSD tradeBins http 429")
				return fmt.Errorf("Margin req http 429: %v", s)
			case s >= 400:
				logger.SendLoggerNotification("ETHUSD tradeBins http >= 400")
				return stop{fmt.Errorf("client error: %v", s)}
			case tradeBins[0].Close == 0:
				logger.SendLoggerNotification("ETHUSD tradeBins[0].Close is 0")
				return fmt.Errorf("ETHUSD tradeBins[0].Close is 0")
			default:
				return nil
			}
		})

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
		Retry(3, 3*time.Second, func() error {
			positions, res, err = f.apiClient.PositionApi.PositionGet(f.auth, &params)
			if res == nil || err != nil {
				println("PositionGet ", err.Error())
				return fmt.Errorf("network error: %v", 1)
			}
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
				logger.SendLoggerNotification("PositionGet http >= 500")
				return fmt.Errorf("server error: %v", s)
			case s == 429:
				time.Sleep(10 * time.Second)
				logger.SendLoggerNotification("PositionGet http 429")
				return fmt.Errorf("Margin req http 429: %v", s)
			case s >= 400:
				logger.SendLoggerNotification("PositionGet http >= 400")
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
		Retry(3, 3*time.Second, func() error {
			margin, res, err = f.apiClient.UserApi.UserGetMargin(f.auth, nil)
			if res == nil || err != nil {
				println("UserGetMargin ", err.Error())
				return fmt.Errorf("network error: %v", 1)
			}

			s := res.StatusCode
			switch {
			case s >= 500:
				logger.SendLoggerNotification("Margin req http >= 500")
				return fmt.Errorf("server error: %v", s)
			case s == 429:
				time.Sleep(10 * time.Second)
				logger.SendLoggerNotification("Margin req http 429")
				return fmt.Errorf("Margin req http 429: %v", s)
			case s >= 400:
				logger.SendLoggerNotification("Margin req http >= 400")
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

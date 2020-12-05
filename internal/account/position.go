package account

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/reposandermets/take-positions/internal/logger"
	"github.com/spf13/viper"
	"github.com/zmxv/bitmexgo"
)

type Flow struct {
	apiClient *bitmexgo.APIClient
	auth      context.Context
}

var F = Flow{}

func (f *Flow) Initialize() {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	var apiKey string
	var apiSecret string
	isTestnet := viper.GetBool("IS_TESTNET")
	if isTestnet {
		apiKey = viper.GetString("API_KEY_TESTNET")
		apiSecret = viper.GetString("API_SECRET_TESTNET")
	} else {
		apiKey = viper.GetString("API_KEY")
		apiSecret = viper.GetString("API_SECRET")
	}

	f.auth = bitmexgo.NewAPIKeyContext(apiKey, apiSecret)

	var netTransport = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 3 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 3 * time.Second,
	}
	var netClient = &http.Client{
		Timeout:   time.Second * 3,
		Transport: netTransport,
	}
	var cfg *bitmexgo.Configuration
	if isTestnet {
		cfg = &bitmexgo.Configuration{
			BasePath:      "https://testnet.bitmex.com/api/v1",
			DefaultHeader: make(map[string]string),
			UserAgent:     "server",
			HTTPClient:    netClient,
		}

		f.apiClient = bitmexgo.NewAPIClient(cfg)
	} else {
		cfg = &bitmexgo.Configuration{
			BasePath:      "https://www.bitmex.com/api/v1",
			DefaultHeader: make(map[string]string),
			UserAgent:     "server",
			HTTPClient:    netClient,
		}
		f.apiClient = bitmexgo.NewAPIClient(cfg)
	}
}

func (f *Flow) HandleQueueItem(payload Payload) {
	var accountState AccountState
	accountState = f.FetchAccountState(payload.Ticker)

	if accountState.MarginError != nil || accountState.PositionError != nil || accountState.TradeBinError != nil || accountState.TradeBinEthError != nil {
		println("ERROR EARLY", accountState.MarginError, accountState.PositionError, accountState.TradeBinError, accountState.TradeBinEthError)
		logger.SendSlackNotification("ERROR EARLY f.FetchAccountState")
		return
	}

	posInfo := "NO open position"
	if accountState.HasOpenPosition {
		posInfo = "HAS open " + accountState.Side + " position."
	}

	shouldClosePosition := accountState.HasOpenPosition && ((payload.Signal == "ExitBuy" && accountState.Side == "Buy") ||
		(payload.Signal == "ExitSell" && accountState.Side == "Sell") ||
		(payload.Signal == "Buy" && accountState.Side == "Sell") ||
		(payload.Signal == "Sell" && accountState.Side == "Buy"))

	if shouldClosePosition {
		logger.SendSlackNotification("INFO about to close position " + accountState.Side + " " + payload.Ticker + " " + posInfo + " Signal: " + payload.Signal + " " + fmt.Sprintf("%d", payload.Sig))

		err := f.ClosePosition(payload.Ticker)
		if err != nil {
			logger.SendSlackNotification("ERROR closing position " + payload.Ticker + " " + err.Error())
		} else {
			logger.SendSlackNotification("INFO OK ClosePosition " + accountState.Side + " " + payload.Ticker)
		}

		err = f.CancelOrders(payload.Ticker)
		if err != nil {
			logger.SendSlackNotification("ERROR CancelOrders " + payload.Ticker + " " + err.Error())
		} else {
			logger.SendSlackNotification("INFO OK CancelOrders" + accountState.Side + " " + payload.Ticker)
		}
		return
	}

	if !accountState.HasOpenPosition && (payload.Signal == "Buy" || payload.Signal == "Sell") {
		logger.SendSlackNotification("INFO " + payload.Ticker + " " + posInfo + " Signal: " + payload.Signal + " " + fmt.Sprintf("%d", payload.Sig))
		f.CancelOrders(payload.Ticker)
		println("Open position ", payload.Signal, " ticker ", payload.Ticker)

		// calculate position size
		positionSize := CalculatePositionSize(accountState, payload)
		println("positionSize ", positionSize)
		logger.SendSlackNotification("positionSize: " + fmt.Sprintf("%d", positionSize))

		if positionSize < 2 {
			logger.SendSlackNotification("ERROR position size: " + fmt.Sprintf("%d", positionSize))

			return
		}

		errOrderMarket := f.OrderMarket(positionSize, payload)
		if errOrderMarket != nil {
			logger.SendSlackNotification("ERROR OrderMarket new Position " + payload.Ticker + " " + errOrderMarket.Error())

			return
		}

		errOpenPosition := Retry(10, 3*time.Second, func() error {
			accountState = f.FetchAccountState(payload.Ticker)

			if accountState.PositionError != nil {
				println("OrderMarket ", accountState.PositionError.Error())
				return fmt.Errorf("Check position after market order network error: %v", 1)
			}

			if !accountState.HasOpenPosition {
				println("ERROR !accountState.HasOpenPosition after marketOrder")
				return fmt.Errorf("no position after marketOrder: %v", 1)
			}

			return nil
		})

		if errOpenPosition != nil {
			println("ERROR FINAL !accountState.HasOpenPosition after marketOrder")
			logger.SendSlackNotification("ERROR ABORT !accountState.HasOpenPosition after marketOrder: " + errOpenPosition.Error())

			return
		}

		logger.SendSlackNotification("INFO Entry price: " + fmt.Sprintf("%f", accountState.Position.AvgEntryPrice))

		sl, tp := CalculateSlTp(accountState, payload)

		err := f.OrderSlTp(accountState, sl, tp)
		if err == nil {
			logger.SendSlackNotification("INFO SUCCESS TP SL")
		} else {
			println("ERROR Market TP SL, about to close position ", err.Error())
			logger.SendSlackNotification("ERROR Market TP SL, FIX manually " + err.Error())
		}
		return
	}
}

package account

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/reposandermets/take-positions/internal/logger"
	"github.com/spf13/viper"
	"github.com/zmxv/bitmexgo"
)

type Flow struct {
	apiClient      *bitmexgo.APIClient
	auth           context.Context
	strategyConfig StrategyConfig
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

	f.strategyConfig = StrategyConfig{
		StepsAllowed:             viper.GetFloat64("STEPS_ALLOWED"),
		LeverageAllowed:          viper.GetFloat64("LEVERAGE_ALLOWED_BUY"),
		LossPercentageForReEntry: viper.GetFloat64("LOSS_PERCENTAGE_FOR_RE_ENTRY"),
	}
}

func (f *Flow) HandleQueueItem(payload Payload) {
	payload.Ticker = strings.Replace(payload.Ticker, "/", "", -1)

	logger.SendSlackNotification("Received signal " + payload.Signal + " " + payload.Ticker)
	if payload.Type != "Active" { // TODO get from ENV
		logger.SendSlackNotification("Signal type mismatch: " + payload.Type)
		return
	}

	var accountState AccountState
	accountState = f.FetchAccountState(payload.Ticker)

	if accountState.MarginError != nil || accountState.PositionError != nil || accountState.TradeBinError != nil || accountState.TradeBinEthError != nil {
		println("ERROR about to retry", accountState.MarginError, accountState.PositionError, accountState.TradeBinError, accountState.TradeBinEthError)
		time.Sleep(3333 * time.Millisecond)
		accountState = f.FetchAccountState(payload.Ticker)
	}

	if accountState.MarginError != nil || accountState.PositionError != nil || accountState.TradeBinError != nil || accountState.TradeBinEthError != nil {
		println("ERROR error getting account data cancel flow", accountState.MarginError, accountState.PositionError, accountState.TradeBinError, accountState.TradeBinEthError)
		logger.SendSlackNotification("ERROR error getting account data cancel flow")
		return
	}

	println("HasOpenPosition ", accountState.HasOpenPosition)
	logger.SendSlackNotification("HasOpenPosition")

	// if no open position cancel all orders
	if !accountState.HasOpenPosition {
		// might want to specify ticker in the future
		f.apiClient.OrderApi.OrderCancelAll(f.auth, nil)
	}

	if accountState.Side != "" && accountState.Side != payload.Signal {
		println("Signal and position side mismatch, position side: ", accountState.Side)
		logger.SendSlackNotification("Cancel flow Signal and position side mismatch, position side: " + accountState.Side)
		return
	}

	accountState.PositionSize, accountState.ProfitPercentage = CalculatePositionSize(accountState, f.strategyConfig, payload)
	if accountState.PositionSize < 0 {
		accountState.PositionSize, accountState.ProfitPercentage = CalculatePositionSize(accountState, f.strategyConfig, payload)
	}

	if accountState.PositionSize < 0 {
		logger.SendSlackNotification("ERROR position size: " + fmt.Sprintf("%d", accountState.PositionSize))

		return
	}

	println("PositionSize ", accountState.PositionSize)

	if accountState.PositionSize > 0 {
		var res *http.Response
		var err error

		_, res, err = f.OrderMarket(accountState, payload)

		if !(err == nil && res.StatusCode >= 200 && res.StatusCode < 300) {
			println("ERROR creating market order RETRY: ", err.Error(), " statuscode ", res.StatusCode)
			time.Sleep(1333 * time.Millisecond)
			_, res, err = f.OrderMarket(accountState, payload)
		}

		if err == nil && res.StatusCode >= 200 && res.StatusCode < 300 {
			time.Sleep(1333 * time.Millisecond)
			accountState = f.FetchAccountState(payload.Ticker)

			if accountState.MarginError != nil || accountState.PositionError != nil || accountState.TradeBinError != nil || accountState.TradeBinEthError != nil {
				println("ERROR accountState after market order about to retry", accountState.MarginError, accountState.PositionError, accountState.TradeBinError, accountState.TradeBinEthError)
				time.Sleep(3333 * time.Millisecond)
				accountState = f.FetchAccountState(payload.Ticker)
			} else if !accountState.HasOpenPosition {
				println("WARNING no open position after market order, about to retry")
				time.Sleep(1333 * time.Millisecond)
				accountState = f.FetchAccountState(payload.Ticker)
			}

			if accountState.MarginError != nil || accountState.PositionError != nil || accountState.TradeBinError != nil || accountState.TradeBinEthError != nil {
				println("ERROR accountState after market order canceling flow", accountState.MarginError.Error(), accountState.PositionError.Error(), accountState.TradeBinError.Error(), accountState.TradeBinEthError.Error())
				logger.SendSlackNotification("ERROR accountState after market order canceling flow")

				return
			}

			// calculate stop loss
			sl, tp, trail := CalculateSlTpTrail(accountState, payload)
			println("SL ", sl)
			println("TP ", tp)
			println("Trail ", trail)
			if sl > 0 || tp > 0 || trail != 0 {
				_, res, err = f.OrderSlTpTrail(accountState, sl, tp, trail)
				if !(err == nil && res.StatusCode >= 200 && res.StatusCode < 300) {
					println("ERROR OrderSlTp RETRY: ", err.Error(), " statuscode ", res.StatusCode)
					time.Sleep(3333 * time.Millisecond)
					_, res, err = f.OrderSlTpTrail(accountState, sl, tp, trail)
				}

				if err == nil && res.StatusCode >= 200 && res.StatusCode < 300 {
					logger.SendSlackNotification("SUCCESS Flow with SL/TP/Trail")
					return
				} else {
					println("ERROR SL/TP: ", err.Error(), " statuscode ", res.StatusCode)
					logger.SendSlackNotification("ERROR SL/TP")
					return
				}
			}

			println("SUCCESS Flow")
		} else {
			println("ERROR creating market order: ", err.Error(), " statuscode ", res.StatusCode)
			logger.SendSlackNotification("ERROR creating market order")
		}
	}
}

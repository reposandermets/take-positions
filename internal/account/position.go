package account

import (
	"context"
	"net/http"
	"strings"
	"time"

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

	// TODO provide improved http client to handle timeouts etc
	if isTestnet {
		f.apiClient = bitmexgo.NewAPIClient(bitmexgo.NewTestnetConfiguration())
	} else {
		f.apiClient = bitmexgo.NewAPIClient(bitmexgo.NewConfiguration())
	}

	f.strategyConfig = StrategyConfig{
		StepsAllowed:             viper.GetFloat64("STEPS_ALLOWED"),
		LeverageAllowed:          viper.GetFloat64("LEVERAGE_ALLOWED_BUY"),
		LossPercentageForReEntry: viper.GetFloat64("LOSS_PERCENTAGE_FOR_RE_ENTRY"),
	}
}

func (f *Flow) HandleQueueItem(payload Payload) {
	payload.Ticker = strings.Replace(payload.Ticker, "/", "", -1)
	println("############################")
	println("Received signal", payload.Signal, " ", payload.Ticker)
	println("############################")

	if payload.Type != "Active" { // TODO get from ENV
		println("Signal type mismatch: ", payload.Type)

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
		println("ERROR canceling flow", accountState.MarginError, accountState.PositionError, accountState.TradeBinError, accountState.TradeBinEthError)

		return
	}

	println("HasOpenPosition ", accountState.HasOpenPosition)

	// if no open position cancel all orders
	if !accountState.HasOpenPosition {
		// might want to specify ticker
		f.apiClient.OrderApi.OrderCancelAll(f.auth, nil)
	}

	if accountState.Side != "" && accountState.Side != payload.Signal {
		println("Signal and position side mismatch, position side: ", accountState.Side)

		return
	}

	accountState.PositionSize, accountState.ProfitPercentage = CalculatePositionSize(accountState, f.strategyConfig, payload)

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
				println("ERROR accountState after market order canceling flow", accountState.MarginError, accountState.PositionError, accountState.TradeBinError, accountState.TradeBinEthError)

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
					println("SUCCESS Flow with SL/TP/Trail")

					return
				} else {
					println("ERROR SL/TP: ", err.Error(), " statuscode ", res.StatusCode)

					return
				}
			}

			println("SUCCESS Flow")
		} else {
			println("ERROR creating market order: ", err.Error(), " statuscode ", res.StatusCode)
		}
	}
}

package account

import (
	"context"

	"github.com/spf13/viper"
	"github.com/zmxv/bitmexgo"
)

type Flow struct {
	apiClient      *bitmexgo.APIClient
	auth           context.Context
	strategyConfig StrategyConfig
}

var F = Flow{}

func (f *Flow) FetchAccountState() (accountState AccountState) {
	cPosition := make(chan PositionState)
	cMargin := make(chan MarginState)
	cTradeBin := make(chan TradeBinState)

	go func() {
		var params bitmexgo.PositionGetOpts
		params.Filter.Set("{\"symbol\":\"XBTUSD\"}")
		positions, _, err := f.apiClient.PositionApi.PositionGet(f.auth, &params)
		var position bitmexgo.Position
		side := ""
		hasOpenPosition := false
		if len(positions) > 0 {
			position = positions[0]
			hasOpenPosition = position.IsOpen
			if position.CurrentQty > 0 {
				side = "Buy"
			} else if position.CurrentQty < 0 {
				side = "Sell"
			}
		}
		cPosition <- PositionState{
			HasOpenPosition: hasOpenPosition,
			Side:            side,
			Position:        position,
			Error:           err,
		}
	}()

	go func() {
		margin, _, err := f.apiClient.UserApi.UserGetMargin(f.auth, nil)
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
		tradeBins, _, err := f.apiClient.TradeApi.TradeGetBucketed(f.auth, &params)
		var tradeBin bitmexgo.TradeBin
		if len(tradeBins) > 0 {
			tradeBin = tradeBins[0]
		}
		cTradeBin <- TradeBinState{
			TradeBin: tradeBin,
			Error:    err,
		}
	}()

	for i := 0; i < 3; i++ {
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
		}
	}

	return accountState
}

func (f *Flow) OrderMarket(accountState AccountState) {
	var params bitmexgo.OrderNewOpts
	params.OrdType.Set("Market")
	params.Side.Set(accountState.Side)
	params.OrderQty.Set(accountState.PositionSize)
	order, _, err := f.apiClient.OrderApi.OrderNew(f.auth, "XBTUSD", &params)
	if err != nil {
		println("OrderMarket error", err.Error())
	} else {
		println("OrderMarket success: ", order.Side, " ", order.OrderQty)
	}
}

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
		LeverageAllowed:          viper.GetFloat64("LEVERAGE_ALLOWED"),
		LossPercentageForReEntry: viper.GetFloat64("LOSS_PERCENTAGE_FOR_RE_ENTRY"),
	}
}

func (f *Flow) HandleQueueItem(payload Payload) {
	println("############################")
	println("Received signal", payload.Signal)
	println("############################")

	accountState := f.FetchAccountState()

	accountState.PositionSize, accountState.ProfitPercentage = CalculatePositionSize(accountState, f.strategyConfig)
	println(accountState.PositionSize)
	if accountState.PositionSize > 0 {
		f.OrderMarket(accountState)
	}
}

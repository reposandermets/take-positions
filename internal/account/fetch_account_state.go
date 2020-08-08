package account

import "github.com/zmxv/bitmexgo"

func (f *Flow) FetchAccountState(symbol string) (accountState AccountState) {
	cPosition := make(chan PositionState)
	cMargin := make(chan MarginState)
	cTradeBin := make(chan TradeBinState)
	cTradeBinEth := make(chan TradeBinEthState)

	go func() {
		var params bitmexgo.PositionGetOpts
		params.Filter.Set("{\"symbol\":\"" + symbol + "\"}")
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
		params.Symbol.Set("ETHUSD")
		params.Reverse.Set(true)
		tradeBins, _, err := f.apiClient.TradeApi.TradeGetBucketed(f.auth, &params)
		var tradeBin bitmexgo.TradeBin
		if len(tradeBins) > 0 {
			tradeBin = tradeBins[0]
		}
		cTradeBinEth <- TradeBinEthState{
			TradeBinEth: tradeBin,
			Error:       err,
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

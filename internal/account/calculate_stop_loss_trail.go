package account

import "math"

func CalculateSlTp(accountState AccountState, payload Payload) (sl float64, tp float64) {
	if payload.Ticker == "XBTUSD" && accountState.Side == "Buy" {
		tp = math.Ceil(accountState.Position.AvgEntryPrice + payload.Atr)
		sl = math.Floor(accountState.Position.AvgEntryPrice - payload.AtrSl)
		return sl, tp
	}

	if payload.Ticker == "XBTUSD" && accountState.Side == "Sell" {
		tp = math.Floor(accountState.Position.AvgEntryPrice - payload.Atr)
		sl = math.Ceil(accountState.Position.AvgEntryPrice + payload.AtrSl)
		return sl, tp
	}

	if payload.Ticker == "ETHUSD" && accountState.Side == "Buy" {
		tp = math.Ceil((accountState.Position.AvgEntryPrice+payload.Atr)*10) / 10
		sl = math.Floor((accountState.Position.AvgEntryPrice-payload.AtrSl)*10) / 10
		return sl, tp
	}

	if payload.Ticker == "ETHUSD" && accountState.Side == "Sell" {
		tp = math.Floor((accountState.Position.AvgEntryPrice-payload.Atr)*10) / 10
		sl = math.Ceil((accountState.Position.AvgEntryPrice+payload.AtrSl)*10) / 10
		return sl, tp
	}

	return sl, tp
}

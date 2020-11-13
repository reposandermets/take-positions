package account

import "math"

func CalculateSlTp(accountState AccountState, payload Payload) (sl float64, tp float64) {
	if payload.Ticker == "ETHUSD" && accountState.Side == "Buy" {
		tp = math.Round((accountState.Position.AvgEntryPrice+payload.Atr)*10) / 10
		sl = math.Round((accountState.Position.AvgEntryPrice-(payload.AtrSl))*10) / 10
		return sl, tp
	}

	if payload.Ticker == "ETHUSD" && accountState.Side == "Sell" {
		tp = math.Round((accountState.Position.AvgEntryPrice-payload.Atr)*10) / 10
		sl = math.Round((accountState.Position.AvgEntryPrice+(payload.AtrSl))*10) / 10
		return sl, tp
	}

	return sl, tp
}

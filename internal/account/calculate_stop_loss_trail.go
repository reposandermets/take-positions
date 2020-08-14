package account

import "math"

func CalculateSlTpTrail(accountState AccountState, payload Payload) (sl float64, tp float64, trail float64) {

	if payload.SlPerc > 0 {
		if payload.Ticker == "ETHUSD" {
			ethPriceRaw := math.Round(accountState.Position.AvgEntryPrice * 100)
			slFactor := ethPriceRaw * payload.SlPerc / 100
			if accountState.Side == "Buy" {
				sl = math.Round((ethPriceRaw-slFactor)/10) / 10
			} else if accountState.Side == "Sell" {
				sl = math.Round((ethPriceRaw+slFactor)/10) / 10
			}
		}
	}

	if payload.TpPerc > 0 {
		if payload.Ticker == "ETHUSD" {
			ethPriceRaw := math.Round(accountState.Position.AvgEntryPrice * 100)
			tpAmount := ethPriceRaw * payload.TpPerc / 100
			if accountState.Side == "Buy" {
				tp = math.Round((ethPriceRaw+tpAmount)/10) / 10
			} else if accountState.Side == "Sell" {
				tp = math.Round((ethPriceRaw-tpAmount)/10) / 10
			}
		}
	}

	if payload.TrailPerc > 0 {
		if payload.Ticker == "ETHUSD" {
			trail = math.Round(accountState.Position.AvgEntryPrice*payload.TrailPerc/10) / 10
			if accountState.Side == "Buy" {
				trail = 0 - trail
			}
		}
	}

	return sl, tp, trail
}

package account

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/reposandermets/take-positions/internal/logger"
)

func FormatFloat(num float64) string {
	prc := 6
	var (
		zero, dot = "0", "."
		str       = fmt.Sprintf("%."+strconv.Itoa(prc)+"f", num)
	)
	return strings.TrimRight(strings.TrimRight(str, zero), dot)
}

func CalculatePositionSize(accountState AccountState, strategyConfig StrategyConfig, payload Payload) (positionSize int) {
	xbtWallet := float64(accountState.Margin.WalletBalance) / 100000000
	atrSl := payload.AtrSl
	equity := xbtWallet * accountState.TradeBin.Close
	riskAllowed := payload.Risk
	positionLeverage := 0.0
	contractsCashValue := 0.0
	if payload.Ticker == "XBTUSD" {
		close := accountState.TradeBin.Close
		atrRiskPerc := atrSl * 100 / close
		riskRatio := riskAllowed / atrRiskPerc
		positionSize = int(math.Floor(xbtWallet * close * riskRatio))
	} else if payload.Ticker == "ETHUSD" {
		close := accountState.TradeBinEth.Close
		atrRiskPerc := atrSl * 100 / close
		riskRatio := riskAllowed / atrRiskPerc
		contractValue := accountState.TradeBinEth.Close / 1000000
		availableContracts := xbtWallet / contractValue
		positionSize = int(math.Floor(availableContracts * riskRatio))
		if positionSize%2 != 0 {
			positionSize = positionSize - 1
		}
		positionLeverage = float64(positionSize) / availableContracts
		contractsCashValue = float64(positionSize) * contractValue * accountState.TradeBin.Close
	}

	logger.SendSlackNotification("INFO equity: " + fmt.Sprintf("%d", equity) +
		" riskAllowed: " + fmt.Sprintf("%d", riskAllowed) +
		" leverage: " + fmt.Sprintf("%d", positionLeverage) +
		" leverage: " + fmt.Sprintf("%d", positionLeverage) +
		" cash: " + fmt.Sprintf("%d", contractsCashValue))

	return positionSize
}

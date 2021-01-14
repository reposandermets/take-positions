package account

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/reposandermets/take-positions/internal/logger"
	"github.com/spf13/viper"
)

func FormatFloat(num float64) string {
	prc := 6
	var (
		zero, dot = "0", "."
		str       = fmt.Sprintf("%."+strconv.Itoa(prc)+"f", num)
	)
	return strings.TrimRight(strings.TrimRight(str, zero), dot)
}

func CalculatePositionSize(accountState AccountState, payload Payload) (positionSize int) {
	xbtWallet := viper.GetFloat64("XBT_WALLET")
	// xbtWallet := float64(accountState.Margin.WalletBalance) / 100000000
	atrSl := payload.AtrSl
	equity := xbtWallet * accountState.TradeBin.Close
	riskAllowed := payload.Risk
	positionLeverage := 0.0
	contractsCashValue := 0.0
	riskRatio := 0.0
	if payload.Ticker == "XBTUSD" {
		close := accountState.TradeBin.Close
		atrRiskPerc := atrSl * 100 / close
		riskRatio = riskAllowed / atrRiskPerc
		positionSize = int(math.Floor(xbtWallet * close * riskRatio))
		positionLeverage = riskRatio
		contractsCashValue = float64(positionSize)
	} else if payload.Ticker == "ETHUSD" {
		close := accountState.TradeBinEth.Close
		atrRiskPerc := atrSl * 100 / close
		riskRatio = riskAllowed / atrRiskPerc
		contractValue := accountState.TradeBinEth.Close / 1000000
		availableContracts := xbtWallet / contractValue
		positionSize = int(math.Floor(availableContracts * riskRatio))
		positionLeverage = float64(positionSize) / availableContracts
		contractsCashValue = float64(positionSize) * contractValue * accountState.TradeBin.Close
	}

	logger.SendLoggerNotification("INFO equity wallet $: " + fmt.Sprintf("%g", math.Floor(equity*100)/100) +
		" wallet XBT: " + fmt.Sprintf("%g", float64(accountState.Margin.WalletBalance)) +
		" riskAllowed: " + fmt.Sprintf("%g", math.Round(riskAllowed*100)/100) +
		" riskRatio: " + fmt.Sprintf("%g", math.Round(riskRatio*100)/100) +
		" leverage: " + fmt.Sprintf("%g", math.Round(positionLeverage*100)/100) +
		" position cash value: " + fmt.Sprintf("%g", math.Round(contractsCashValue*100)/100))

	return positionSize
}

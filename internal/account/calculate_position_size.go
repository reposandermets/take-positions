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

	leverageRequiredForStep := strategyConfig.LeverageAllowed
	leverageAvailable := strategyConfig.LeverageAllowed - accountState.Margin.MarginLeverage

	hasNotEnoughLeverageLeft := leverageRequiredForStep > leverageAvailable // && math.Abs(profitPercentage-strategyConfig.LossPercentageForReEntry) > 1e-6

	logger.SendSlackNotification("ETHUSD leverageRequiredForStep: " + fmt.Sprintf("%f", leverageRequiredForStep))
	if hasNotEnoughLeverageLeft {
		logger.SendSlackNotification("Not enough levarage left. available: " + FormatFloat(leverageAvailable) + " required: " + FormatFloat(leverageRequiredForStep))

		return positionSize
	}

	xbtWallet := float64(accountState.Margin.WalletBalance) / 100000000
	logger.SendSlackNotification("ETHUSD accountState.Margin.WalletBalance: " + fmt.Sprintf("%d", accountState.Margin.WalletBalance))
	if payload.Ticker == "XBTUSD" {
		positionSize = int(math.Floor(xbtWallet * accountState.TradeBin.Close * leverageRequiredForStep))
	} else if payload.Ticker == "ETHUSD" {
		logger.SendSlackNotification("ETHUSD accountState.TradeBinEth.Close: " + fmt.Sprintf("%f", accountState.TradeBinEth.Close))
		contractValue := accountState.TradeBinEth.Close / 1000000
		println("ETHUSD contract value: ", contractValue)
		availableContracts := xbtWallet / contractValue
		println("ETHUSD available contracts: ", availableContracts)
		logger.SendSlackNotification("ETHUSD available contracts: " + fmt.Sprintf("%f", availableContracts))
		positionSize = int(math.Floor(availableContracts * leverageRequiredForStep))
	}

	if positionSize%2 != 0 {
		positionSize = positionSize - 1
	}

	return positionSize
}

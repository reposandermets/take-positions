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

func CalculatePositionSize(accountState AccountState, strategyConfig StrategyConfig, payload Payload) (positionSize int, profitPercentage float64) {

	leverageRequiredForStep := strategyConfig.LeverageAllowed / strategyConfig.StepsAllowed
	leverageAvailable := strategyConfig.LeverageAllowed - accountState.Margin.MarginLeverage

	hasNotEnoughLeverageLeft := leverageRequiredForStep > leverageAvailable &&
		math.Abs(profitPercentage-strategyConfig.LossPercentageForReEntry) > 1e-6

	logger.SendSlackNotification("ETHUSD leverageRequiredForStep: " + fmt.Sprintf("%f", leverageRequiredForStep))
	if hasNotEnoughLeverageLeft {
		logger.SendSlackNotification("Not enough levarage left. available: " + FormatFloat(leverageAvailable) + " required: " + FormatFloat(leverageRequiredForStep))

		return positionSize, profitPercentage
	}

	if accountState.HasOpenPosition {
		if accountState.Side == "Buy" {
			profitPercentage = (accountState.TradeBin.Close/accountState.Position.AvgEntryPrice - 1) * 100
		}

		if accountState.Side == "Sell" {
			profitPercentage = (1 - accountState.TradeBin.Close/accountState.Position.AvgEntryPrice) * 100
		}

		if profitPercentage > strategyConfig.LossPercentageForReEntry {

			println(profitPercentage > strategyConfig.LossPercentageForReEntry)
			println("profitPercentage: ", profitPercentage)
			println("strategyConfig.LossPercentageForReEntry: ", strategyConfig.LossPercentageForReEntry)
			println("profitPercentage > strategyConfig.LossPercentageForReEntry, no reentry")

			return positionSize, profitPercentage
		}
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

	return positionSize, profitPercentage
}

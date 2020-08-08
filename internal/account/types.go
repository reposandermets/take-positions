package account

import (
	"github.com/zmxv/bitmexgo"
)

type Payload struct {
	Ticker   string  `json:"ticker"`
	Exchange string  `json:"exchange"`
	Signal   string  `json:"signal"`
	Type     string  `json:"type"`
	SlPerc   float64 `json:"sl_perc"`
	TpPerc   float64 `json:"tp_perc"`
}

type AccountState struct {
	PositionSize     int
	ProfitPercentage float64
	Side             string
	HasOpenPosition  bool
	Position         bitmexgo.Position
	PositionError    error
	Margin           bitmexgo.Margin
	MarginError      error
	TradeBin         bitmexgo.TradeBin
	TradeBinError    error
	TradeBinEth      bitmexgo.TradeBin
	TradeBinEthError error
}

type PositionState struct {
	Side            string
	HasOpenPosition bool
	Position        bitmexgo.Position
	Error           error
}

type MarginState struct {
	Margin bitmexgo.Margin
	Error  error
}

type TradeBinState struct {
	TradeBin bitmexgo.TradeBin
	Error    error
}

type TradeBinEthState struct {
	TradeBinEth bitmexgo.TradeBin
	Error       error
}

type StrategyConfig struct {
	StepsAllowed             float64
	LeverageAllowed          float64
	LossPercentageForReEntry float64
}

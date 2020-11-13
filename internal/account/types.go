package account

import (
	"github.com/zmxv/bitmexgo"
)

type Payload struct {
	Sig      int     `json:"sig"`
	Ticker   string  `json:"ticker"`
	Atr      float64 `json:"atr"`
	AtrSl    float64 `json:"atrsl"`
	Risk     float64 `json:"risk"`
	Type     string  `json:"type"`
	Secret   string  `json:"secret"`
	Exchange string  `json:"exchange"`
	Signal   string  `json:"signal"`
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
	LeverageAllowed float64
}

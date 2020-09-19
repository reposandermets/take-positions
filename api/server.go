package api

import (
	"github.com/reposandermets/take-positions/api/controllers"
	"github.com/spf13/viper"
)

var server = controllers.Server{}

func Run() {

	println("STEPS_ALLOWED: ", int(viper.GetFloat64("STEPS_ALLOWED")))
	println("LEVERAGE_ALLOWED_BUY: ", int(viper.GetFloat64("LEVERAGE_ALLOWED_BUY")))
	println("LOSS_PERCENTAGE_FOR_RE_ENTRY: ", viper.GetFloat64("LOSS_PERCENTAGE_FOR_RE_ENTRY"))
	println("IS_TESTNET: ", viper.GetBool("IS_TESTNET"))
	server.Initialize()
	server.Run(":8080")
}

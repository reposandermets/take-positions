package api

import (
	"github.com/reposandermets/take-positions/api/controllers"
	"github.com/spf13/viper"
)

var server = controllers.Server{}

func Run() {
	println("IS_TESTNET: ", viper.GetBool("IS_TESTNET"))
	server.Initialize()
	server.Run(":8080")
}

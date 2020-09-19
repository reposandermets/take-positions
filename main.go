package main

import (
	"github.com/reposandermets/take-positions/api"
	core "github.com/reposandermets/take-positions/internal"
	"github.com/reposandermets/take-positions/internal/logger"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()
	logger.SendSlackNotification("Boot")
	core.Run()
	api.Run()
}

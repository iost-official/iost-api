package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func ReadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("$GOPATH/src/github.com/iost-official/iost-api/config")
	//viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

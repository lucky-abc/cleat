package config

import (
	"fmt"
	"github.com/spf13/viper"
)

var config *viper.Viper

func InitSystemConfig(file string, configPath string) {
	config = viper.New()
	config.SetConfigName(file)
	config.SetConfigType("yaml")

	config.AddConfigPath(configPath)
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
}

func Config() *viper.Viper {
	return config
}

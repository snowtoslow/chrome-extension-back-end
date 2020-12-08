package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

func Init() error {

	viper.SetConfigType("yaml")

	// specify where the config file is
	viper.SetConfigFile("/home/snowtoslow/go/src/chrome-extension-back-end/config/config.yml") // or viper.SetConfigFile("config/config"), doesn't matter

	// read in config file and check your errors
	if err := viper.ReadInConfig(); err != nil {
		// handle errors
		log.Println("ERR:", err)
	}

	// confirm where the file has been read in from
	fmt.Println(viper.ConfigFileUsed())

	return viper.ReadInConfig()
}

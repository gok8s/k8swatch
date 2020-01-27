package utils

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

func InitConfig() {
	//box := packr.NewBox("../configs")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	} else {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Search config in home directory with name ".eventwatch" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath("/etc/k8swatch/configs")
		viper.AddConfigPath("/configs/")
		viper.AddConfigPath("./configs")
		viper.SetConfigType("yaml")

		config := os.Getenv("K8SWATCH_CONF")
		env := os.Getenv("K8SWATCH_ENV")

		if config != "" {
			fmt.Printf("trying to load diy configfile %s\n", config)
			viper.SetConfigName(config)
			if err := viper.ReadInConfig(); err != nil {
				fmt.Printf("load config file: %s error:%v\n", config, err)
				return
			}
			fmt.Printf("load config file:%v success\n", viper.ConfigFileUsed())
			return
		} else if env != "" {
			fmt.Printf("trying to load environment %s config\n", env)
			viper.SetConfigName(env)
			if err := viper.ReadInConfig(); err != nil {
				fmt.Printf("load config file:%v error:%v\n", viper.ConfigFileUsed(), err)
				return
			}
			fmt.Printf("enable env config file:%v success\n", viper.ConfigFileUsed())
		} else {
			fmt.Println("trying to load default config")
			viper.SetConfigName("default")
			if err := viper.ReadInConfig(); err != nil {
				fmt.Printf("load default config file error:%v\n", err)
				return
			}
			fmt.Printf("load default config success")
		}

	}
	viper.AutomaticEnv() // read in environment variables that match
}

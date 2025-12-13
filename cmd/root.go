package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	envFile string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "futures-trading",
	Short:   "futures-trading service",
	Long:    "futures-trading service",
	Version: "0.0.0",
}

// SetVersion inject version from git
func SetVersion(r string) {
	if len(r) > 0 {
		RootCmd.Version = r
	}
	viper.SetDefault("service_version", RootCmd.Version)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&envFile, "env", ".env", "env file (default is .env)")
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "config file (default is config.yaml)")
}

func initConfig() {
	// Load .env file if exists
	if _, err := os.Stat(envFile); err == nil {
		if err := godotenv.Load(envFile); err != nil {
			log.Fatalf("Error loading env file %s: %v", envFile, err)
		} else {
			fmt.Println("Loaded env file:", envFile)
		}
	} else {
		fmt.Println("No env file found, skipping:", envFile)
	}

	// Load YAML config file
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Cannot read config file: %s", err)
	} else {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("__", "."))
	viper.AutomaticEnv() // This will override config with ENV variables
}

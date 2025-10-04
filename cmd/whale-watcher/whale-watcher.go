package main

import (
	"fmt"
	"os"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/docs"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/validator"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfg config.Config
var cfgFile string

var envPrefix = "WHALE_WATCHER_"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "whale-watcher",
	Short: "Your way to watch your containers",
	Long:  `Enforce best practices across your application and check Dockerfiles and container for compliance`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.Unmarshal(&cfg); err != nil {
			return err
		}
		fmt.Printf("Loaded config:\n%+v\n", cfg)
		return nil
	},
}

func main() {
	rootCmd.AddCommand(docs.NewCommand())

	rootCmd.AddCommand(validator.NewCommand())

	// Add flag/env for config file itself
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to config file (default: ./config.yaml)")
	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindEnv("config", fmt.Sprintf("%sCONFIG_FILE", envPrefix))

	// Dynamically add flags/envs for all config fields
	if err := config.AddConfigFlags(rootCmd, "", &cfg, envPrefix); err != nil {
		panic(fmt.Sprintf("failed to add config flags: %v", err))
	}

	// Config file loading
	cobra.OnInitialize(func() {
		file := viper.GetString("config")
		if file == "" {
			file = "./config.yaml"
		}
		viper.SetConfigFile(file)
		viper.SetConfigType("yaml")

		if err := viper.ReadInConfig(); err == nil {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}
	})

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

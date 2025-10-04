package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/config"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/docs"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/validator"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// 1️⃣ Determine config file path
		cfgFile := viper.GetString("config")
		if cfgFile == "" {
			cfgFile = "./config.yaml"
		}

		viper.SetConfigFile(cfgFile)
		viper.SetConfigType("yaml")

		// 2️⃣ Try to read the config file
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				fmt.Println("No config file found, using defaults/env/flags")
			} else {
				return fmt.Errorf("failed to read config file: %w", err)
			}
		} else {
			fmt.Println("Using config file:", viper.ConfigFileUsed())
		}

		// 3️⃣ Environment variable setup
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		// 4️⃣ Bind all command flags (so flags override config/env)
		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return fmt.Errorf("failed to bind flags: %w", err)
		}
		if err := viper.BindPFlags(cmd.PersistentFlags()); err != nil {
			return fmt.Errorf("failed to bind persistent flags: %w", err)
		}

		return nil
	}}

func setHelpFuncRecursive(cmd *cobra.Command, helpFunc func(*cobra.Command, []string)) {
	cmd.SetHelpFunc(helpFunc)
	for _, sub := range cmd.Commands() {
		setHelpFuncRecursive(sub, helpFunc)
	}
}

func helpFunctionOverride(cmd *cobra.Command, args []string) {
	fmt.Println(cmd.Short)
	fmt.Println("\nUsage:")
	fmt.Printf("  %s [flags]\n\n", cmd.Use)

	// print grouped flags
	grouped := map[string][]*pflag.Flag{}

	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		group := f.Annotations["group"]
		groupName := "Global Options"
		if len(group) > 0 {
			groupName = group[0]
		}
		grouped[groupName] = append(grouped[groupName], f)
	})

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		group := f.Annotations["group"]
		groupName := "Global Options"
		if len(group) > 0 {
			groupName = group[0]
		}
		grouped[groupName] = append(grouped[groupName], f)
	})

	for groupName, flags := range grouped {
		fmt.Printf("\n%s:\n", groupName)
		for _, f := range flags {
			fmt.Printf("  --%-20s %s\n", f.Name, f.Usage)
		}
	}
}

func main() {
	// Add flag/env for config file itself
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Path to config file (default: ./config.yaml)")
	_ = viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	_ = viper.BindEnv("config", fmt.Sprintf("%sCONFIG_FILE", envPrefix))

	// Dynamically add flags/envs for all config fields
	if err := config.AddConfigFlagsWithGroups(rootCmd, "", &cfg, envPrefix); err != nil {
		panic(fmt.Sprintf("failed to add config flags: %v", err))
	}

	setHelpFuncRecursive(rootCmd, helpFunctionOverride)

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout}).With().Logger()

	rootCmd.AddCommand(docs.NewCommand())
	rootCmd.AddCommand(validator.NewCommand())
	rootCmd.AddCommand(config.NewCommand())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

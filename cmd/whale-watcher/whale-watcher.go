/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"os"

	"github.com/coffeemakingtoaster/whale-watcher/pkg/docs"
	"github.com/coffeemakingtoaster/whale-watcher/pkg/validator"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "whale-watcher",
	Short: "Your way to watch your containers",
	Long:  `Enforce best practices across your application and check Dockerfiles and container for compliance`,
}

func main() {
	rootCmd.AddCommand(docs.NewCommand())
	rootCmd.AddCommand(validator.NewCommand())

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

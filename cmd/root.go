package cmd

import (
	"github.com/bdlm/log"
	"github.com/spf13/cobra"
)

var (
	timestamps bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use: "thrempp",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		// workaround for Fatal (os.Exit(1))
		//TODO https://github.com/stretchr/testify/issues/813
		log.Panic(err)
	}
}

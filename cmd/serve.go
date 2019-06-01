package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/bdlm/log"
	"github.com/bdlm/std/logger"
	"github.com/spf13/cobra"

	"dev.sum7.eu/genofire/golang-lib/database"
	"dev.sum7.eu/genofire/golang-lib/file"

	"dev.sum7.eu/genofire/thrempp/component"
	// need for database init
	_ "dev.sum7.eu/genofire/thrempp/component/all"
	_ "dev.sum7.eu/genofire/thrempp/models"
)

type Config struct {
	LogLevel   logger.Level       `toml:"log_level"`
	Database   database.Config    `toml:"database"`
	Components []component.Config `toml:"component"`
}

var configPath string

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:     "serve",
	Short:   "Run xmpp transport",
	Example: "yanic serve --config /etc/thrempp.toml",
	Run: func(cmd *cobra.Command, args []string) {
		config := &Config{}
		if err := file.ReadTOML(configPath, config); err != nil {
			log.Panicf("open config file: %s", err)
		}

		log.SetLevel(config.LogLevel)

		if err := database.Open(config.Database); err != nil {
			log.Panicf("no database connection: %s", err)
		}
		defer database.Close()
		component.Load(config.Components)

		// Wait for INT/TERM
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigs
		log.Infof("received %s", sig)
	},
}

func init() {
	RootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&configPath, "config", "c", "config.toml", "Path to configuration file")
}

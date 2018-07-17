// nolint: errcheck
package main

import (
	"strings"

	"github.com/lawrencegripper/ion/internal/app/handler"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var handlerConfig = handler.NewConfiguration()

var handlerCmdConfig = viper.New()

// NewHandlerCommand create the handler command with its flags
func NewHandlerCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "ion-handler",
		Short: "ion-handler to embed in the processing",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			handlerConfig.LogFile = handlerCmdConfig.GetString("log-file")
			handlerConfig.LogLevel = handlerCmdConfig.GetString("log-level")
			handlerConfig.PrintConfig, _ = cmd.Flags().GetBool("printconfig")

			if handlerCmdConfig.GetBool("development.enabled") {
				handlerConfig.DevelopmentConfiguration.Enabled = true
				handlerConfig.DevelopmentConfiguration.BaseDir = handlerCmdConfig.GetString("development.basedir")
			}

			// Globally set configuration level
			switch strings.ToLower(handlerConfig.LogLevel) {
			case "debug":
				log.SetLevel(log.DebugLevel)
			case "info":
				log.SetLevel(log.InfoLevel)
			case "warn":
				log.SetLevel(log.WarnLevel)
			case "error":
				log.SetLevel(log.ErrorLevel)
			default:
				log.SetLevel(log.WarnLevel)
			}

			return nil
		},
	}

	cmd.AddCommand(NewStartCommand())
	cmd.AddCommand(NewVersionCommand())

	flags := cmd.PersistentFlags()

	flags.StringP("logfile", "L", "", "File to log output to")
	flags.StringP("loglevel", "l", "warn", "Logging level, possible values {debug, info, warn, error}")
	flags.Bool("development.enabled", false, "A flag to enable development features")
	flags.String("development.basedir", ".dev", "Base directory to write development output to")
	flags.BoolP("printconfig", "P", false, "Set to print config on start")

	handlerCmdConfig.BindPFlag("log-file", flags.Lookup("logfile"))
	handlerCmdConfig.BindPFlag("log-level", flags.Lookup("loglevel"))
	handlerCmdConfig.BindPFlag("development.enabled", flags.Lookup("development.enabled"))
	handlerCmdConfig.BindPFlag("development.basedir", flags.Lookup("development.basedir"))

	return cmd
}

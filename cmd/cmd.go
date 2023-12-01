package cmd

import (
	"github.com/leaktk/scanner/pkg/config"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

const cliLong = `Name:
  leaktk-scanner - Search for secrets

Description:
  The leaktk scanner is meant to run in two primary modes. There is the "scan"
  mode for single ad-hoc scans, and the "listen" mode for streaming scan
  requests as JSON lines. The listen mode is best for bulk scans or when
  integrating the scanner in other services.
`

const configDescription = `config file path
order of precedence:
1. --config/-c
2. env var LEAKTK_CONFIG
3. ${XDG_CONFIG_HOME}/leatktk/config.toml
4. /etc/leatktk/config.toml
5. The default config
`

func runHelp(cmd *cobra.Command, args []string) {
	cmd.Help()
}

func initLog() {
	// TODO
}

// NewCommand provides a built Command for the app to use
func rootCommand() *cobra.Command {
	cobra.OnInitialize(initLog)

	rootCommand := &cobra.Command{
		Use:   "leaktk-scanner",
		Short: "The scanner component in LeakTK",
		Long:  cliLong,
		Run:   runHelp,
	}

	flags := rootCommand.PersistentFlags()
	flags.StringP("config", "c", "", configDescription)

	return rootCommand
}

// Execute the command and parse the args
func Execute() {
	if err := rootCommand().Execute(); err != nil {
		if strings.Contains(err.Error(), "unknown flag") {
			os.Exit(config.ExitCodeBlockingError)
		}
		log.Fatal(err.Error())
	}
}

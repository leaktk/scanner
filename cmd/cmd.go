package cmd

import (
	"fmt"
	"github.com/leaktk/scanner/pkg/config"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strings"
)

// Version number set by the build
var Version = ""

// Commit id set by the build
var Commit = ""

const configDescription = `config file path

order of precedence:

1. --config/-c
2. env var LEAKTK_CONFIG
3. ${XDG_CONFIG_HOME}/leatktk/config.toml
4. /etc/leatktk/config.toml
5. The default config
`
const scanKindDescription = `what kind of resource is being scanned

supported values:

- GitRepo: If this is provided, resource should be something git can clone or
  a file path to scan the repo
- Path: If this is provided, resource should be a local file path that should
  be scanned
`

const scanOptionDescription = `additional options to pass to the scan
format:

--option key=value --option key=value ...

The supported values depend on the kind of scan. If multiple of the same key
are used, the value will become a list.

Check out the README for supported options:

https://github.com/leaktk/scanner/blob/main/README.md

Note: You may want to run 'leaktk-scanner version' to make sure the README
aligns with the version you're using.
`

func runHelp(cmd *cobra.Command, args []string) {
	if err := cmd.Help(); err != nil {
		log.Fatal(err.Error())
	}
}

func initLogger() {
	// TODO
}

func runLogin(cmd *cobra.Command, args []string) {}
func loginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Log into a pattern server",
		Run:   runLogin,
	}
}

func runScan(cmd *cobra.Command, args []string) {
	fmt.Println("TODO")
}

func scanCommand() *cobra.Command {
	scanCommand := &cobra.Command{
		Use:   "scan",
		Short: "Preform ad-hoc scans",
		Run:   runScan,
	}

	flags := scanCommand.PersistentFlags()
	flags.String("kind", "GitRepo", scanKindDescription)
	flags.String("id", "", "an ID for tying responses to requests")
	flags.String("resource", "", "what will be scanned (what goes here depends on kind)")
	flags.StringArray("option", []string{}, scanOptionDescription)

	return scanCommand
}

func runListen(cmd *cobra.Command, args []string) {
	fmt.Println("TODO")
}

func listenCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "listen",
		Short: "Listen for scan requests on stdin",
		Run:   runListen,
	}
}

func runVersion(cmd *cobra.Command, args []string) {
	if len(Version) > 0 {
		fmt.Printf("Version: %v\n", Version)

		if len(Commit) > 0 {
			fmt.Printf("Commit: %v\n", Commit)
		}
	} else {
		fmt.Println("Version information not available")
	}
}

func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display the scanner version",
		Run:   runVersion,
	}
}

// NewCommand provides a built Command for the app to use
func rootCommand() *cobra.Command {
	cobra.OnInitialize(initLogger)

	rootCommand := &cobra.Command{
		Use:   "leaktk-scanner",
		Short: "LeakTK Scanner: Scan for secrets",
		Run:   runHelp,
	}

	flags := rootCommand.PersistentFlags()
	flags.StringP("config", "c", "", configDescription)

	rootCommand.AddCommand(loginCommand())
	rootCommand.AddCommand(scanCommand())
	rootCommand.AddCommand(listenCommand())
	rootCommand.AddCommand(versionCommand())

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

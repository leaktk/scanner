package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"sync"

	"github.com/leaktk/scanner/pkg/config"
	"github.com/leaktk/scanner/pkg/id"
	"github.com/leaktk/scanner/pkg/logger"
	"github.com/leaktk/scanner/pkg/scanner"
)

const maxRequestSize = 256 * 1024

// Version number set by the build
var Version = ""

// Commit id set by the build
var Commit = ""

const scanKindDescription = `what kind of resource is being scanned

supported values:

- GitRepo: If this is provided, resource should be something git can clone or
  a file path to scan the repo

- Path: If this is provided, resource should be a local file path that should
  be scanned

`

const scanOptionsDescription = `additional options to pass to the scan
format:

--options '{json...}'

Check out the README for supported options:

https://github.com/leaktk/scanner/blob/main/README.md

Note: You may want to run 'leaktk-scanner version' to make sure the README
aligns with the version you're using.

`

var cfg *config.Config

func initLogger() {
	if err := logger.SetLoggerLevel("INFO"); err != nil {
		logger.Warning("could not set log level to INFO")
	}
}

func runHelp(cmd *cobra.Command, args []string) {
	if err := cmd.Help(); err != nil {
		logger.Fatal("%v", err)
	}
}

func runLogin(cmd *cobra.Command, args []string) {
	logger.Info("logging in: pattern_server=%q", cfg.Scanner.Patterns.Server.URL)

	fmt.Print("Enter auth token: ")

	var authToken string
	if _, err := fmt.Scanln(&authToken); err != nil {
		logger.Fatal("could not login: error=%q", err)
	}

	if err := config.SavePatternServerAuthToken(authToken); err != nil {
		logger.Fatal("could not login: error=%q", err)
	}

	logger.Info("token saved")
}

func runLogout(cmd *cobra.Command, args []string) {
	logger.Info("logging out: pattern_server=%q", cfg.Scanner.Patterns.Server.URL)

	if err := config.RemovePatternServerAuthToken(); err != nil {
		logger.Fatal("could not logout: error=%q", err)
	}

	logger.Info("token removed")
}

func loginCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Log into a pattern server",
		Run:   runLogin,
	}
}

func logoutCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out of a pattern server",
		Run:   runLogout,
	}
}

func runScan(cmd *cobra.Command, args []string) {
	request, err := scanCommandToRequest(cmd)

	if err != nil {
		logger.Fatal(err.Error())
	}

	leakScanner := scanner.NewScanner(cfg)
	defer leakScanner.Close()

	leakScanner.Send(request)
	fmt.Println(<-leakScanner.Responses())
}

func scanCommandToRequest(cmd *cobra.Command) (*scanner.Request, error) {
	flags := cmd.Flags()

	id, err := flags.GetString("id")
	if err != nil || len(id) == 0 {
		return nil, errors.New("missing required field: id")
	}

	kind, err := flags.GetString("kind")
	if err != nil || len(kind) == 0 {
		return nil, errors.New("missing required field: kind")
	}

	resource, err := flags.GetString("resource")
	if err != nil || len(resource) == 0 {
		return nil, errors.New("missing required field: resource")
	}

	rawOptions, err := flags.GetString("options")
	if err != nil {
		return nil, fmt.Errorf("there was an issue with the options flag (%v)", err)
	}

	options := make(map[string]any)
	if err := json.Unmarshal([]byte(rawOptions), &options); err != nil {
		return nil, fmt.Errorf("could not parse options (%v)", err)
	}

	requestData, err := json.Marshal(
		map[string]any{
			"id":       id,
			"kind":     kind,
			"resource": resource,
			"options":  options,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %v", err)
	}

	var request scanner.Request

	err = json.Unmarshal(requestData, &request)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %v", err)
	}

	return &request, nil
}

func scanCommand() *cobra.Command {
	scanCommand := &cobra.Command{
		Use:   "scan",
		Short: "Preform ad-hoc scans",
		Run:   runScan,
	}

	flags := scanCommand.Flags()
	flags.StringP("id", "", id.ID(), "an ID for tying responses to requests")
	flags.StringP("kind", "k", "GitRepo", scanKindDescription)
	flags.StringP("resource", "r", "", "what will be scanned (what goes here depends on kind)")
	flags.StringP("options", "o", "{}", scanOptionsDescription)

	return scanCommand
}

func runListen(cmd *cobra.Command, args []string) {
	var wg sync.WaitGroup

	stdinScanner := bufio.NewScanner(os.Stdin)
	stdinScanner.Buffer(make([]byte, maxRequestSize), maxRequestSize)
	leakScanner := scanner.NewScanner(cfg)
	defer leakScanner.Close()

	// Prints the output of the scanner as they come
	go func() {
		for response := range leakScanner.Responses() {
			fmt.Println(response)
			wg.Done()
		}
	}()

	// Listen for requests
	for stdinScanner.Scan() {
		var request scanner.Request
		err := json.Unmarshal(stdinScanner.Bytes(), &request)

		if err != nil {
			logger.Error("%v: request_id=%q", err, request.ID)
			continue
		}

		if len(request.Resource.String()) == 0 {
			logger.Error("no resource provided: request_id=%q", request.ID)
			continue
		}

		wg.Add(1)
		leakScanner.Send(&request)
	}

	if err := stdinScanner.Err(); err != nil {
		logger.Error("%v", err)
	}

	wg.Wait()
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

func configure(cmd *cobra.Command, args []string) error {
	switch cmd.Use {
	case "listen":
		if err := logger.SetLoggerFormat(logger.JSON); err != nil {
			return err
		}
	default:
		if err := logger.SetLoggerFormat(logger.HUMAN); err != nil {
			return err
		}
	}
	path, err := cmd.Flags().GetString("config")

	if err == nil {
		// If path == "", this will look other places
		cfg, err = config.LocateAndLoadConfig(path)

		if err == nil {
			err = logger.SetLoggerLevel(cfg.Logger.Level)
		}
	}

	return err
}

func rootCommand() *cobra.Command {
	cobra.OnInitialize(initLogger)

	rootCommand := &cobra.Command{
		Use:               "leaktk-scanner",
		Short:             "LeakTK Scanner: Scan for secrets",
		Run:               runHelp,
		PersistentPreRunE: configure,
	}

	flags := rootCommand.PersistentFlags()
	flags.StringP("config", "c", "", "config file path")

	rootCommand.AddCommand(loginCommand())
	rootCommand.AddCommand(logoutCommand())
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
		logger.Fatal("%v", err.Error())
	}
}

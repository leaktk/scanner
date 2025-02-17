package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/leaktk/scanner/pkg/response"

	"github.com/spf13/cobra"

	"github.com/leaktk/scanner/pkg/config"
	"github.com/leaktk/scanner/pkg/id"
	"github.com/leaktk/scanner/pkg/logger"
	"github.com/leaktk/scanner/pkg/scanner"
)

// Version number set by the build
var Version = ""

// Commit id set by the build
var Commit = ""

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
		logger.Fatal("could not generate scan request: error=%q", err.Error())
	}

	formatter, _ := response.NewFormatter(cfg.Formatter)

	var wg sync.WaitGroup
	leakScanner := scanner.NewScanner(cfg)

	// Prints the output of the scanner as they come
	go leakScanner.Recv(func(response *response.Response) {
		fmt.Println(formatter.Format(response))
		wg.Done()
	})

	wg.Add(1)
	leakScanner.Send(request)

	wg.Wait()
}

func scanCommandToRequest(cmd *cobra.Command) (*scanner.Request, error) {
	flags := cmd.Flags()

	id, err := flags.GetString("id")
	if err != nil || len(id) == 0 {
		return nil, errors.New("missing required field: field=\"id\"")
	}

	kind, err := flags.GetString("kind")
	if err != nil || len(kind) == 0 {
		return nil, errors.New("missing required field: field=\"kind\"")
	}

	resource, err := flags.GetString("resource")
	if err != nil || len(resource) == 0 {
		return nil, errors.New("missing required field: field=\"resource\"")
	}

	rawOptions, err := flags.GetString("options")
	if err != nil {
		return nil, fmt.Errorf("there was an issue with the options flag: error=%q", err)
	}

	options := make(map[string]any)
	if err := json.Unmarshal([]byte(rawOptions), &options); err != nil {
		return nil, fmt.Errorf("could not parse options: error=%q", err)
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
		return nil, fmt.Errorf("could not marshal requestData: error=%q", err)
	}

	var request scanner.Request

	err = json.Unmarshal(requestData, &request)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal requestData: error=%q", err)
	}

	return &request, nil
}

func scanCommand() *cobra.Command {
	scanCommand := &cobra.Command{
		Use:   "scan",
		Short: "Perform ad-hoc scans",
		Run:   runScan,
	}

	flags := scanCommand.Flags()
	flags.StringP("id", "", id.ID(), "an ID for tying responses to requests")
	flags.StringP("kind", "k", "GitRepo", "the kind of resource being scanned")
	flags.StringP("resource", "r", "", "what will be scanned (what goes here depends on kind)")
	flags.StringP("options", "o", "{}", "additional request options formatted as JSON")

	return scanCommand
}

func readLine(reader *bufio.Reader) ([]byte, error) {
	var buf bytes.Buffer

	for {
		line, isPrefix, err := reader.ReadLine()
		buf.Write(line)

		if err != nil || !isPrefix {
			return buf.Bytes(), err
		}
	}
}

func runListen(cmd *cobra.Command, args []string) {
	var wg sync.WaitGroup

	stdinReader := bufio.NewReader(os.Stdin)
	leakScanner := scanner.NewScanner(cfg)

	// Prints the output of the scanner as they come
	go leakScanner.Recv(func(response *response.Response) {
		fmt.Println(response)
		wg.Done()
	})

	// Listen for requests
	for {
		line, err := readLine(stdinReader)

		if err != nil {
			if err == io.EOF {
				break
			}

			logger.Error("error reading from stdin: error=%q", err)
			continue
		}

		var request scanner.Request
		err = json.Unmarshal(line, &request)

		if err != nil {
			logger.Error("could not unmarshal request: error=%q", err)
			continue
		}

		if len(request.Resource.String()) == 0 {
			logger.Error("no resource provided: request_id=%q", request.ID)
			continue
		}

		wg.Add(1)
		leakScanner.Send(&request)
	}

	// Wait for all of the scans to complete and responses to be sent
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
		if err != nil {
			return err
		}
	}

	// If a format is specified on the command line update the application config.
	format, err := cmd.Flags().GetString("format")
	if err == nil && format != "" {
		cfg.Formatter.Format = format
	}

	// Check if the OutputFormat is valid
	_, err = response.GetOutputFormat(cfg.Formatter.Format)
	if err != nil {
		logger.Fatal("%v", err.Error())
	}

	truncate, err := cmd.Flags().GetInt("truncate")
	if err == nil {
		cfg.Formatter.Truncate = truncate
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
	flags.StringP("format", "f", "", "output format [json(default), human, csv, toml, yaml]")
	flags.IntP("truncate", "t", -1, "Specify length to truncate result secret, matches and context to, -1 no truncation")

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

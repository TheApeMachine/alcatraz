package cmd

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/theapemachine/alcatraz"
	"github.com/theapemachine/errnie"

	"github.com/theapemachine/alcatraz/pkg/config"
)

/*
Embed a mini filesystem into the binary to hold the default config file.
This will be written to the home directory of the user running the service,
which allows a developer to easily override the config file.
*/
//go:embed cfg/config.yml
var embedded embed.FS

/*
rootCmd represents the base command when called without any subcommands
*/
var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "alcatraz",
		Short: "Alcatraz is a tool for managing containerized environments",
		Long:  rootLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			environment := alcatraz.NewEnvironment(cmd.Context(), args[0])

			switch args[1] {
			case "start":
				return environment.Start()
			case "stop":
				return environment.Stop()
			case "destroy":
				return environment.Close()
			}

			return nil
		},
	}
)

/*
Execute adds all child commands to the root command and sets flags appropriately.
This is called by main.main(). It only needs to happen once to the rootCmd.
*/
func Execute() {
	if errnie.Error(rootCmd.Execute()) != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(
		&cfgFile,
		"config",
		"",
		"path to config file (default: try cmd/cfg/config.yml, ./config.yml, $HOME/.alcatraz/config.yml, then embedded default)",
	)
}

/*
initConfig loads cfg/config.yml from, in order:
  - path given by --config (if set)
  - ./cmd/cfg/config.yml (repo checkout)
  - ./config.yml
  - $HOME/.alcatraz/config.yml
  - embedded cfg/config.yml
*/
func initConfig() {
	errnie.Apply(config.NewErrnieConfig().ToLibraryConfig())

	viper.SetConfigType("yml")

	tryRead := func(path string) error {
		viper.SetConfigFile(path)
		return viper.ReadInConfig()
	}

	loaded := false

	if rootCmd.PersistentFlags().Changed("config") && strings.TrimSpace(cfgFile) != "" {
		readResult := errnie.Does(func() (any, error) {
			return nil, tryRead(cfgFile)
		})

		if readResult.Err() == nil {
			loaded = true
		} else {
			fmt.Fprintf(os.Stderr, "alcatraz: config file %q: %v\n", cfgFile, readResult.Err())
			os.Exit(1)
		}
	}

	if !loaded {
		paths := []string{
			"cmd/cfg/config.yml",
			"config.yml",
		}

		homeResult := errnie.Does(os.UserHomeDir)

		if homeResult.Err() == nil {
			paths = append(paths, filepath.Join(homeResult.Value(), ".alcatraz", "config.yml"))
		}

		for _, path := range paths {
			readResult := errnie.Does(func() (any, error) {
				return nil, tryRead(path)
			})

			if readResult.Err() == nil {
				loaded = true
				break
			}
		}
	}

	if !loaded {
		openResult := errnie.Does(func() (fs.File, error) {
			return embedded.Open("cfg/config.yml")
		})

		if openResult.Err() != nil {
			fmt.Printf("embedded config file not found: %v\n", openResult.Err())
			return
		}

		cfgReader := openResult.Value()
		defer cfgReader.Close()

		readResult := errnie.Does(func() (any, error) {
			return nil, viper.ReadConfig(cfgReader)
		})

		if readResult.Err() != nil {
			fmt.Printf("embedded config file not readable: %v\n", readResult.Err())
			return
		}
	}

	viper.WatchConfig()
}

const rootLong = `
Alcatraz is a tool for managing containerized environments.
It was designed to give A.I. agents a full, safe Linux environment.
`

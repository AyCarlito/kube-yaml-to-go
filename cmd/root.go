package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/AyCarlito/kube-yaml-to-go/pkg/logger"
)

func init() {
	rootCmd.PersistentFlags().StringVar(&inputFilePath, "input", "", "Path to input file. Reads from stdin if unset.")
	rootCmd.PersistentFlags().StringVar(&outputFilePath, "output", "", "Path to output file. Writes to stdout if unset.")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enables verbose output. Generates a full source file.")
}

// CLI Flags
var (
	inputFilePath  string
	outputFilePath string
	verbose        bool
)

var rootCmd = &cobra.Command{
	Use:           "kube-yaml-to-go",
	Short:         "Generate Go source code from Kubernetes YAML.",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Build logger.
		log, err := logger.NewZapConfig().Build()
		if err != nil {
			panic(fmt.Errorf("failed to build zap logger: %v", err))
		}
		cmd.SetContext(logger.ContextWithLogger(cmd.Context(), log))
		cmd.Parent().SetContext(logger.ContextWithLogger(cmd.Parent().Context(), log))
		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		// By default, cobra prints the error and usage string on every error.
		// We only desire this behaviour in the case where command line parsing fails e.g. unknown command or flag.
		// Cobra does not provide a mechanism for achieving this fine grain control, so we implement our own.
		if strings.Contains(err.Error(), "command") || strings.Contains(err.Error(), "flag") {
			// Parsing errors are printed along with the usage string.
			fmt.Println(err.Error())
			fmt.Println(rootCmd.UsageString())
		} else {
			// Other errors logged, no usage string displayed.
			log := logger.LoggerFromContext(rootCmd.Context())
			log.Error(err.Error())
		}
		os.Exit(1)
	}
}

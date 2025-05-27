package cmd

import (
	"github.com/spf13/cobra"

	"github.com/AyCarlito/kube-yaml-to-go/pkg/generator"
)

func init() {
	rootCmd.AddCommand(generateCmd)
}

// generateCmd is the command for generating Go source code from Kubernetes YAML files.
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates Go source code.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return generator.NewGenerator(cmd.Context(), inputFilePath, outputFilePath, verbose).Generate()
	},
}

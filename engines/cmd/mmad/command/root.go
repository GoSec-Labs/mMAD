package command

import (
	"fmt"
	"mMMAD/engines/internal/config"

	"github.com/spf13/cobra"
)

var (
	cfg     *config.Config
	verbose bool
	output  string
)

var rootCmd = &cobra.Command{
	Use:     "mmad",
	Short:   "mMAD is a is a production-grade stablecoin management system with  zero-knowledge proofs for privacy-preserving reserve verification.",
	Version: "0.1.0",
}

func Execute(config *config.Config) error {
	cfg = config
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "text", "Output format (text, json, yaml)")

	// Add subcommands
	rootCmd.AddCommand(proofCmd)
	rootCmd.AddCommand(reserveCmd)
	//rootCmd.AddCommand(complianceCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("mMAD CLI v%s\n", cfg.App.Version)
		fmt.Print("Environment: %s\n", cfg.App.Environment)
	},
}

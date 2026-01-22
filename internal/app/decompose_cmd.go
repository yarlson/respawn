package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	prdPath string
)

var decomposeCmd = &cobra.Command{
	Use:   "decompose",
	Short: "Decompose a PRD into tasks",
	Long:  `Decompose takes a PRD file and breaks it down into actionable tasks in .respawn/tasks.yaml.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Cobra handles required flags, but we can add extra validation if needed.
		fmt.Printf("Decomposing PRD: %s\n", prdPath)
		// Access global flags if needed: GlobalBackend, etc.
		return nil
	},
}

func init() {
	rootCmd.AddCommand(decomposeCmd)

	decomposeCmd.Flags().StringVar(&prdPath, "prd", "", "Path to the PRD file (required)")
	_ = decomposeCmd.MarkFlagRequired("prd")
}

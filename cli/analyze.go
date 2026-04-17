package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tbl-editor/editor"
)

var (
	analyzeTmpDir string
	analyzeMapFile string
	analyzeCmd = &cobra.Command{
		Use:   "analyze",
		Short: "Analyze exported CSVs against map.json and show sample data per field",
		Long:  `Reads CSV files from --tmp and map.json from --map, then prints sample values for every mapped SQL column so you can verify the mappings before seeding.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := editor.Analyze(analyzeTmpDir, analyzeMapFile); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	analyzeCmd.Flags().StringVar(&analyzeTmpDir, "tmp", "tmp", "Directory containing exported CSV files")
	analyzeCmd.Flags().StringVar(&analyzeMapFile, "map", "map.json", "Path to map.json")

	rootCmd.AddCommand(analyzeCmd)
}

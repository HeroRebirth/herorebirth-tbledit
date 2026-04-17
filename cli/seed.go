package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tbl-editor/editor"
)

var (
	seedTmpDir  string
	seedMapFile string
	seedOutFile string
	seedCmd     = &cobra.Command{
		Use:   "seed",
		Short: "Generate SQL INSERT statements from exported CSVs",
		Long:  `Reads CSV files from --tmp and map.json from --map, then writes INSERT ... ON DUPLICATE KEY UPDATE statements to --out.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := editor.SeedSQL(seedTmpDir, seedMapFile, seedOutFile); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		},
	}
)

func init() {
	seedCmd.Flags().StringVar(&seedTmpDir, "tmp", "tmp", "Directory containing exported CSV files")
	seedCmd.Flags().StringVar(&seedMapFile, "map", "map.json", "Path to map.json")
	seedCmd.Flags().StringVarP(&seedOutFile, "out", "o", "herorebirth_seed.sql", "Output SQL file")

	rootCmd.AddCommand(seedCmd)
}

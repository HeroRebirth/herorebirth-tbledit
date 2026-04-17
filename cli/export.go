package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"tbl-editor/editor"
)

var (
	inFile    string
	outFile   string
	dirPath   string
	tmpDir    string
	exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Exports .tbl file(s) to Excel or CSV",
		Long:  `Export a single .tbl to Excel (-i/-o) or all .tbl files in a directory to CSV files in ./tmp (-d).`,
		Run: func(cmd *cobra.Command, args []string) {

			if dirPath != "" {
				// Directory mode: export all .tbl files to CSV
				if err := editor.ExportDir(dirPath, tmpDir); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				return
			}

			if inFile == "" || outFile == "" {
				cmd.Usage()
				return
			}

			editor.Export(inFile, outFile)
		},
	}
)

func init() {
	exportCmd.Flags().StringVarP(&inFile, "input", "i", "", "Input .tbl file")
	exportCmd.Flags().StringVarP(&outFile, "output", "o", "", "Output .xlsx file")
	exportCmd.Flags().StringVarP(&dirPath, "dir", "d", "", "Directory of .tbl files to export as CSVs")
	exportCmd.Flags().StringVar(&tmpDir, "tmp", "tmp", "Output directory for CSV files (default: ./tmp)")

	rootCmd.AddCommand(exportCmd)
}

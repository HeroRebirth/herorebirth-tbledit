package editor

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ExportDir reads all .tbl files from srcDir, parses them, and writes
// one CSV per file into dstDir (created if it does not exist).
// CSV column headers are "col_0", "col_1", ... matching the .tbl column order.
func ExportDir(srcDir, dstDir string) error {
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	entries, err := filepath.Glob(filepath.Join(srcDir, "*.tbl"))
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return fmt.Errorf("no .tbl files found in %s", srcDir)
	}

	ok, failed := 0, 0
	for _, tblPath := range entries {
		base := filepath.Base(tblPath)
		csvName := strings.TrimSuffix(base, filepath.Ext(base)) + ".csv"
		csvPath := filepath.Join(dstDir, csvName)

		if err := exportTBLtoCSV(tblPath, csvPath); err != nil {
			fmt.Fprintf(os.Stderr, "SKIP %s: %v\n", base, err)
			failed++
		} else {
			fmt.Printf("OK   %s -> %s\n", base, csvName)
			ok++
		}
	}

	fmt.Printf("\nExported %d files, %d failed → %s\n", ok, failed, dstDir)
	return nil
}

func exportTBLtoCSV(tblPath, csvPath string) error {
	tbl, err := ParseTBL(tblPath)
	if err != nil {
		return err
	}

	f, err := os.Create(csvPath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)

	// header row: col_0, col_1, ...
	header := make([]string, len(tbl.ColTypes))
	for i, ct := range tbl.ColTypes {
		header[i] = fmt.Sprintf("col_%d_%s", i, typeTitles[ct])
	}
	if err := w.Write(header); err != nil {
		return err
	}

	for _, row := range tbl.Rows {
		if err := w.Write(row); err != nil {
			return err
		}
	}

	w.Flush()
	return w.Error()
}

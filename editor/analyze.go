package editor

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Analyze reads tmp/ CSVs + map.json and prints sample data per mapped column.
func Analyze(tmpDir, mapFile string) error {
	cfg, err := LoadMapConfig(mapFile)
	if err != nil {
		return err
	}

	for fname, tblCfg := range cfg {
		csvName := strings.TrimSuffix(fname, ".tbl") + ".csv"
		csvPath := filepath.Join(tmpDir, csvName)

		rows, header, err := readCSV(csvPath)
		if err != nil {
			fmt.Printf("\n[%s] MISSING CSV (%v) — run 'tbledit export -d <dir>' first\n", fname, err)
			continue
		}

		fmt.Printf("\n╔══ %s (%d rows) ══\n", fname, len(rows)-1)

		for _, tgt := range tblCfg.Targets {
			fmt.Printf("║  → SQL table: %s\n", tgt.Table)

			// Collect sample values for each SQL field
			fieldSamples := make(map[string][]string)
			fieldOrder := []string{}

			for colIdx, cm := range tgt.Cols {
				if cm.Field == "" {
					continue
				}
				if colIdx >= len(header) {
					fmt.Printf("║    WARNING: col_%d out of range (file has %d cols)\n", colIdx, len(header))
					continue
				}

				key := cm.Field
				if cm.Role != "direct" && cm.Role != "" && cm.Role != "bool" && cm.Role != "float" && cm.Role != "json_text" {
					key = fmt.Sprintf("%s[%s]", cm.Field, cm.Role)
				}

				if _, seen := fieldSamples[key]; !seen {
					fieldOrder = append(fieldOrder, key)
				}

				// Gather up to 3 sample values from the CSV (skip header + skip_col rows)
				startRow := 1 + tgt.SkipCol
				samples := fieldSamples[key]
				for r := startRow; r < len(rows) && len(samples) < 3; r++ {
					if colIdx < len(rows[r]) {
						v := rows[r][colIdx]
						if v != "" && v != "0" && v != "0.0" && v != "null" && v != "NULL" {
							samples = append(samples, v)
						}
					}
				}
				// If all zeros, show the first actual value anyway
				if len(samples) == 0 && len(rows) > startRow && colIdx < len(rows[startRow]) {
					samples = []string{rows[startRow][colIdx]}
				}
				fieldSamples[key] = samples
			}

			for _, key := range fieldOrder {
				samples := fieldSamples[key]
				fmt.Printf("║    %-35s %s\n", key+":", strings.Join(samples, " | "))
			}
		}
		fmt.Println("╚" + strings.Repeat("═", 60))
	}

	return nil
}

func readCSV(path string) ([][]string, []string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	var rows [][]string
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		rows = append(rows, rec)
	}
	if len(rows) == 0 {
		return nil, nil, fmt.Errorf("empty CSV")
	}
	return rows, rows[0], nil
}

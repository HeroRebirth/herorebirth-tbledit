package editor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// ColMapping describes how one TBL column maps to a SQL field.
// It can be either:
//   - a plain string (field name, role="direct")
//   - nil (skip this column)
//   - a ColSpec object with field + role
type ColMapping struct {
	Field      string // SQL field name; empty = skip
	Role       string // "direct","bool","float","array","coord_x","coord_y","len_prefix","mat_key","mat_val","json_text"
	FilterZero bool   // for "array": omit zero values
}

// TableTarget is one SQL table this TBL file feeds into.
type TableTarget struct {
	Table   string       // SQL table name
	SkipCol int          // skip first N data rows (e.g. header row)
	Cols    []ColMapping // one entry per TBL column
}

// TBLConfig holds all targets for one .tbl file.
type TBLConfig struct {
	Targets []TableTarget
}

// MapConfig is the full map.json structure.
type MapConfig map[string]TBLConfig // key = .tbl filename (no path)

// LoadMapConfig reads map.json from the given path.
func LoadMapConfig(path string) (MapConfig, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read map.json: %w", err)
	}

	// map.json uses a flexible ColMapping representation:
	//   null            → skip (Field="")
	//   "fieldname"     → direct mapping
	//   {"field":"...", "role":"...", "filter_zero":true}
	var rawMap map[string]struct {
		Targets []struct {
			Table   string        `json:"table"`
			SkipCol int           `json:"skip_col"`
			Cols    []interface{} `json:"cols"`
		} `json:"targets"`
	}

	if err := json.Unmarshal(raw, &rawMap); err != nil {
		return nil, fmt.Errorf("parse map.json: %w", err)
	}

	result := make(MapConfig)
	for fname, cfg := range rawMap {
		tc := TBLConfig{}
		for _, tgt := range cfg.Targets {
			tt := TableTarget{
				Table:   tgt.Table,
				SkipCol: tgt.SkipCol,
				Cols:    make([]ColMapping, len(tgt.Cols)),
			}
			for i, raw := range tgt.Cols {
				switch v := raw.(type) {
				case nil:
					// skip
				case string:
					tt.Cols[i] = ColMapping{Field: v, Role: "direct"}
				case map[string]interface{}:
					cm := ColMapping{Role: "direct"}
					if f, ok := v["field"].(string); ok {
						cm.Field = f
					}
					if r, ok := v["role"].(string); ok {
						cm.Role = r
					}
					if fz, ok := v["filter_zero"].(bool); ok {
						cm.FilterZero = fz
					}
					tt.Cols[i] = cm
				default:
					return nil, fmt.Errorf("%s col %d: unexpected type %T", fname, i, raw)
				}
			}
			tc.Targets = append(tc.Targets, tt)
		}
		result[fname] = tc
	}

	return result, nil
}

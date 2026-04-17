package editor

import (
	"encoding/binary"
	"errors"
	"io/ioutil"
	"math"
	"strconv"
)

// TBLFile holds the parsed contents of a .tbl binary file.
type TBLFile struct {
	ColTypes []ColType
	Rows     [][]string // each cell is a string representation of the value
}

// errEOF is returned internally to signal a clean stop at end-of-file.
var errEOF = errors.New("eof")

// ParseTBL reads a .tbl binary file and returns all complete rows.
//
// Binary format (little-endian):
//   col_count  UINT32
//   col_type[] UINT32 × col_count
//   row_count  UINT32
//   data_size  UINT32   (= file_size – 4 – col_count*4; spans row_count + row data)
//   row data …
//
// STRING parsing rule: if the column immediately before a STRING column is UINT32
// AND that UINT32 value is ≤ 65535 (a reasonable string length), that value IS
// the byte length and the STRING field has no embedded prefix in the data.
// Otherwise STRING has its own 4-byte embedded length prefix.
//
// Trailing padding: after each row in files that contain at least one STRING
// column, 4 zero bytes are present and are skipped automatically.
//
// Partial last row: many game files store row_count rows but truncate the last
// one by a few bytes. Incomplete rows are silently dropped.
func ParseTBL(path string) (*TBLFile, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	offset := 0
	fileLen := len(data)

	readU32 := func() (uint32, bool) {
		if offset+4 > fileLen {
			return 0, false
		}
		v := binary.LittleEndian.Uint32(data[offset:])
		offset += 4
		return v, true
	}
	readU16 := func() (uint16, bool) {
		if offset+2 > fileLen {
			return 0, false
		}
		v := binary.LittleEndian.Uint16(data[offset:])
		offset += 2
		return v, true
	}
	readU64 := func() (uint64, bool) {
		if offset+8 > fileLen {
			return 0, false
		}
		v := binary.LittleEndian.Uint64(data[offset:])
		offset += 8
		return v, true
	}
	readByte := func() (byte, bool) {
		if offset >= fileLen {
			return 0, false
		}
		v := data[offset]
		offset++
		return v, true
	}
	readBytes := func(n int) ([]byte, bool) {
		if n < 0 || offset+n > fileLen {
			return nil, false
		}
		b := data[offset : offset+n]
		offset += n
		return b, true
	}

	// --- Header ---
	colCount, ok := readU32()
	if !ok {
		return nil, errEOF
	}
	colTypes := make([]ColType, colCount)
	for i := uint32(0); i < colCount; i++ {
		t, ok := readU32()
		if !ok {
			return nil, errEOF
		}
		colTypes[i] = ColType(t)
	}
	rowCount, ok := readU32()
	if !ok {
		return nil, errEOF
	}
	readU32() // data_size — consumed but not used for bounds checking

	hasString := false
	for _, ct := range colTypes {
		if ct == STRING {
			hasString = true
			break
		}
	}

	rows := make([][]string, 0, rowCount)

	for r := uint32(0); r < rowCount; r++ {
		rowStart := offset
		row := make([]string, colCount)
		aborted := false

		var prevU32 uint32
		var prevType ColType

		for i, ct := range colTypes {
			switch ct {
			case BYTE:
				v, ok := readByte()
				if !ok {
					aborted = true
					break
				}
				row[i] = strconv.FormatUint(uint64(v), 10)
				prevType = ct

			case INT16:
				v, ok := readU16()
				if !ok {
					aborted = true
					break
				}
				row[i] = strconv.FormatInt(int64(int16(v)), 10)
				prevType = ct

			case UINT16:
				v, ok := readU16()
				if !ok {
					aborted = true
					break
				}
				row[i] = strconv.FormatUint(uint64(v), 10)
				prevType = ct

			case INT32:
				v, ok := readU32()
				if !ok {
					aborted = true
					break
				}
				row[i] = strconv.FormatInt(int64(int32(v)), 10)
				prevType = ct

			case UINT32:
				v, ok := readU32()
				if !ok {
					aborted = true
					break
				}
				row[i] = strconv.FormatUint(uint64(v), 10)
				prevU32 = v
				prevType = ct

			case UINT64:
				v, ok := readU64()
				if !ok {
					aborted = true
					break
				}
				row[i] = strconv.FormatUint(v, 10)
				prevType = ct

			case FLOAT:
				raw, ok := readU32()
				if !ok {
					aborted = true
					break
				}
				f := math.Float32frombits(raw)
				row[i] = strconv.FormatFloat(float64(f), 'f', -1, 32)
				prevType = ct

			case STRING:
				var slen uint32
				// Use preceding UINT32 as length ONLY if it looks like a valid
				// string length (≤ 65535). Otherwise read an embedded 4-byte prefix.
				if prevType == UINT32 && prevU32 <= 65535 {
					slen = prevU32
				} else {
					v, ok := readU32()
					if !ok {
						aborted = true
						break
					}
					slen = v
				}
				raw, ok := readBytes(int(slen))
				if !ok {
					aborted = true
					break
				}
				row[i] = decodeString(raw)
				prevType = ct

			default:
				// Unknown type — treat as fatal (corrupt file)
				_ = rowStart
				return &TBLFile{ColTypes: colTypes, Rows: rows}, nil
			}

			if aborted {
				break
			}

			if ct != UINT32 {
				prevU32 = 0
			}
		}

		if aborted {
			// Partial last row — stop here
			break
		}

		// Skip 4-byte zero padding after rows in string-containing files
		if hasString && offset+4 <= fileLen {
			if data[offset] == 0 && data[offset+1] == 0 &&
				data[offset+2] == 0 && data[offset+3] == 0 {
				offset += 4
			}
		}

		rows = append(rows, row)
	}

	return &TBLFile{ColTypes: colTypes, Rows: rows}, nil
}

// decodeString converts raw bytes to a printable UTF-8 string.
// Non-ASCII bytes are replaced with '?' (CSV-safe).
func decodeString(raw []byte) string {
	s := make([]byte, 0, len(raw))
	for _, b := range raw {
		if b < 128 {
			s = append(s, b)
		} else {
			s = append(s, '?')
		}
	}
	return string(s)
}

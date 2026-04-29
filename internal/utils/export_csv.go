package utils

import (
	"bytes"
	"encoding/csv"
)

// WriteCSV — generic CSV writer
func WriteCSV(headers []string, rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	// BOM for Excel UTF-8 compatibility
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	w := csv.NewWriter(&buf)
	if err := w.Write(headers); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

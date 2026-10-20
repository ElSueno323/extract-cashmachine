package shop

import (
	"bytes"
	"encoding/csv"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding/charmap"
)

func ParseCSV(data []byte) ([][]string, error) {
	text := decode(data)
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return [][]string{}, nil
	}
	if strings.HasPrefix(trimmed, "<") {
		return nil, errNotCSV
	}

	reader := csv.NewReader(strings.NewReader(text))
	reader.Comma = ';'
	reader.FieldsPerRecord = -1
	reader.LazyQuotes = true
	records, err := reader.ReadAll()
	if err != nil {
		return nil, errNotCSV
	}

	rows := make([][]string, 0, len(records))
	for _, record := range records {
		if rowIsEmpty(record) {
			continue
		}
		rows = append(rows, record)
	}
	return rows, nil
}

func decode(data []byte) string {
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if utf8.Valid(data) {
		return string(data)
	}
	decoded, err := charmap.Windows1252.NewDecoder().Bytes(data)
	if err != nil {
		return string(data)
	}
	return string(decoded)
}

func rowIsEmpty(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

package sheet

import (
	"archive/zip"
	"bytes"

	"github.com/xuri/excelize/v2"
)

func Transactions(rows [][]string) ([]byte, error) {
	file := excelize.NewFile()
	defer file.Close()

	const name = "Transactions"
	if err := file.SetSheetName("Sheet1", name); err != nil {
		return nil, err
	}

	for r, row := range rows {
		for c, value := range row {
			cell, err := excelize.CoordinatesToCellName(c+1, r+1)
			if err != nil {
				return nil, err
			}
			if err := file.SetCellValue(name, cell, value); err != nil {
				return nil, err
			}
		}
	}

	if len(rows) > 0 {
		bold, err := file.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
		if err != nil {
			return nil, err
		}
		if err := file.SetRowStyle(name, 1, 1, bold); err != nil {
			return nil, err
		}
		if err := file.SetPanes(name, &excelize.Panes{
			Freeze:      true,
			YSplit:      1,
			TopLeftCell: "A2",
			ActivePane:  "bottomLeft",
		}); err != nil {
			return nil, err
		}
		setColumnWidths(file, name, rows)
	}

	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func setColumnWidths(file *excelize.File, name string, rows [][]string) {
	columns := 0
	sample := rows
	if len(sample) > 200 {
		sample = sample[:200]
	}
	for _, row := range sample {
		if len(row) > columns {
			columns = len(row)
		}
	}
	for c := 0; c < columns; c++ {
		width := 10.0
		for _, row := range sample {
			if c < len(row) {
				size := float64(len(row[c]) + 2)
				if size > width {
					width = size
				}
			}
		}
		if width > 48 {
			width = 48
		}
		column, err := excelize.ColumnNumberToName(c + 1)
		if err != nil {
			continue
		}
		_ = file.SetColWidth(name, column, column, width)
	}
}

type ZipEntry struct {
	Name string
	Data []byte
}

func Zip(entries []ZipEntry) ([]byte, error) {
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, entry := range entries {
		file, err := writer.Create(entry.Name)
		if err != nil {
			_ = writer.Close()
			return nil, err
		}
		if _, err := file.Write(entry.Data); err != nil {
			_ = writer.Close()
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func TransactionCount(rows [][]string) int {
	if len(rows) == 0 {
		return 0
	}
	return len(rows) - 1
}

func Preview(rows [][]string, limit int) (headers []string, body [][]string, truncated bool) {
	headers = []string{}
	body = [][]string{}
	if len(rows) == 0 {
		return headers, body, false
	}
	headers = append(headers, rows[0]...)
	rest := rows[1:]
	if limit > 0 && len(rest) > limit {
		return headers, append([][]string{}, rest[:limit]...), true
	}
	return headers, rest, false
}

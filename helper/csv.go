package helper

import (
	"bytes"
	"encoding/csv"
)

func SanitizeCSVCell(value string) string {
	if value == "" {
		return value
	}

	switch value[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + value
	}

	return value
}

func BuildCSV(header []string, rows [][]string) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("\xEF\xBB\xBF")

	writer := csv.NewWriter(&buf)

	if err := writer.Write(header); err != nil {
		return nil, err
	}

	if err := writer.WriteAll(rows); err != nil {
		return nil, err
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
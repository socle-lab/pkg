package csv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

func ImportCSV[T any](filePath string) ([]T, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("cannot read header: %w", err)
	}

	// Build CSV tag → field index mapping
	var model T
	t := reflect.TypeOf(model)
	if t.Kind() != reflect.Struct {
		return nil, errors.New("generic type must be a struct")
	}

	tagMap := map[int]int{} // csvIndex → structFieldIndex
	for i, col := range header {
		csvCol := strings.TrimSpace(strings.ToLower(col))
		for j := 0; j < t.NumField(); j++ {
			tag := t.Field(j).Tag.Get("csv")
			if strings.ToLower(tag) == csvCol {
				tagMap[i] = j
				break
			}
		}
	}

	var results []T
	// Process rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading row: %w", err)
		}

		val := reflect.New(t).Elem()
		for colIndex, fieldIndex := range tagMap {
			if colIndex >= len(row) {
				continue
			}
			val.Field(fieldIndex).SetString(row[colIndex])
		}

		results = append(results, val.Interface().(T))
	}

	return results, nil
}

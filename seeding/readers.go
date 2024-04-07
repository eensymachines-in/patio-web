package seeding

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

var (
	JsonSeedReader = func(fp string) (SeedReader, error) {
		file, err := os.Open(fp) // inits the reader in the constructor
		if err != nil {
			return nil, err
		}
		return &jsonSeedReader{jsonF: file}, nil
	}
)

type SeedReader interface {
	Read() ([]map[string]interface{}, error)
}

type jsonSeedReader struct {
	jsonF *os.File
}

// reads the json file to byt ready for unmarshalling
func (jsr *jsonSeedReader) Read() ([]map[string]interface{}, error) {
	if jsr.jsonF == nil {
		return nil, fmt.Errorf("nil json, cannot read")
	}
	byt, err := io.ReadAll(jsr.jsonF)
	if err != nil {
		return nil, err
	}
	// data is owned by the writer
	result := []map[string]interface{}{} // irrespective of the type of the data
	if err := json.Unmarshal(byt, &result); err != nil {
		return nil, err
	}
	return result, nil
}

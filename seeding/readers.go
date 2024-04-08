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
	jsonF *os.File // file pointer to the input json file
}

// Read scans through a json file and picks up data to be unmarshalled to map[string]interface{}. Will error in case unable to read file, or the json file pointer is nil. Use JsonSeedReader constructor to first initiliaze the reader.
func (jsr *jsonSeedReader) Read() ([]map[string]interface{}, error) {
	if jsr.jsonF == nil {
		return nil, fmt.Errorf("nil json, cannot read")
	}
	byt, err := io.ReadAll(jsr.jsonF)
	if err != nil {
		return nil, err
	}
	// data is owned by the writer
	// data type is irrespective of the type of the data that gets pushed in the database
	result := []map[string]interface{}{}
	if err := json.Unmarshal(byt, &result); err != nil {
		return nil, err
	}
	return result, nil
}

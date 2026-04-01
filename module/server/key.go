package server

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type KeyChecker interface {
	Check(name string, key string) bool
}

type JSONKeyChecker struct {
	data map[string]string
}

func (this JSONKeyChecker) Check(name string, key string) bool {
	return (this.data[name] == key)
}

func GetJSONKeyChecker() (*JSONKeyChecker, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	jsonPath := filepath.Join(wd, "key.json")
	jsonFile, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	data := make(map[string]string)
	err = json.Unmarshal(jsonFile, &data)
	if err != nil {
		return nil, err
	}

	return &JSONKeyChecker{
		data,
	}, nil
}

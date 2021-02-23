package jsonutil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/iancoleman/orderedmap"
)

func ReadFileBytesOrPanic(filename string) []byte {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("error reading file %q: %v", filename, err))
	}

	return raw
}

func ReadFileOrPanic(filename string) *orderedmap.OrderedMap {
	raw := ReadFileBytesOrPanic(filename)
	converted, err := Unmarshal(raw)
	if err != nil {
		panic(err)
	}
	return converted
}

func Unmarshal(jsonStr []byte) (*orderedmap.OrderedMap, error) {
	var converted orderedmap.OrderedMap
	if err := json.Unmarshal(jsonStr, &converted); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}
	return &converted, nil
}

package jsonutil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func ReadFileBytesOrPanic(filename string) []byte {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("error reading file %q: %v", filename, err))
	}

	return raw
}

func ReadFileOrPanic(filename string) map[string]interface{} {
	raw := ReadFileBytesOrPanic(filename)
	converted, err := Unmarshal(raw)
	if err != nil {
		panic(err)
	}
	return converted
}

func Unmarshal(jsonStr []byte) (map[string]interface{}, error) {
	var converted map[string]interface{}
	if err := json.Unmarshal(jsonStr, &converted); err != nil {
		return nil, fmt.Errorf("invalid JSON input: %w", err)
	}
	return converted, nil
}

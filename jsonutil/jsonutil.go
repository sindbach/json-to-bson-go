package jsonutil

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func ReadFileBytes(filename string) []byte {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Sprintf("error reading file %q: %v", filename, err))
	}

	return raw
}

func ReadFile(filename string) map[string]interface{} {
	raw := ReadFileBytes(filename)
	var unmarshalled map[string]interface{}
	if err := json.Unmarshal(raw, &unmarshalled); err != nil {
		panic(fmt.Sprintf("error unmarshalling JSON: %v", err))
	}

	return unmarshalled
}

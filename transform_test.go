package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func readFile(filename string) []byte {
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return raw
}

func readTestInputFile(filename string) map[string]interface{} {
	raw := readFile(filename)
	var dat map[string]interface{}
	if err := json.Unmarshal(raw, &dat); err != nil {
		panic(err)
	}
	return dat
}

func TestTransformJSON(t *testing.T) {
	input := readTestInputFile("./examples/example_01.json")
	output := readFile("./examples/example_01.go")
	result, err := TransformJSON(input)
	if err != nil {
		t.Errorf("Failed to transform: %s", err)
	}
	//TODO
	fmt.Println(input)
	fmt.Println(string(output))
	fmt.Println(result)
}

package simplejson

import (
	"fmt"
	"testing"

	"github.com/sindbach/json-to-bson-go/jsonutil"
)

func TestTransformJSON(t *testing.T) {
	input := jsonutil.ReadFileBytesOrPanic("./examples/example_01.json")
	output := jsonutil.ReadFileBytesOrPanic("./examples/example_01.go")
	result, err := Convert(input)
	if err != nil {
		t.Errorf("Failed to transform: %s", err)
	}
	//TODO
	fmt.Println(input)
	fmt.Println(string(output))
	fmt.Println(result)
}

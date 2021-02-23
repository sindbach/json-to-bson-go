package extjson

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sindbach/json-to-bson-go/jsonutil"
)

func TestConvert(t *testing.T) {
	testCases := []struct {
		name       string
		inputfile  string
		outputfile string
		expectErr  bool
	}{
		{"scalar values", "extjson_scalar", "simplejson_scalar", false},
		{"primitive values", "extjson_primitive", "primitive", false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := jsonutil.ReadFileBytesOrPanic(fmt.Sprintf("../testdata/%s.json", tc.inputfile))
			actual, err := Convert(input, false)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			expected := jsonutil.ReadFileBytesOrPanic(fmt.Sprintf("../testdata/%s.generated", tc.outputfile))
			if diff := cmp.Diff(string(expected), actual); diff != "" {
				t.Fatalf("Generated struct doesn't match expected, got difference %s", diff)
			}
		})
	}
}

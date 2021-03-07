package convert

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sindbach/json-to-bson-go/jsonutil"
	"github.com/sindbach/json-to-bson-go/options"
)

func TestConvert(t *testing.T) {
	//customNameOpts := options.NewOptions().SetStructName("CustomStruct")
	//minimizeTrueOpts := options.NewOptions().SetMinimizeIntegerSize(true)
	//minimizeFalseOpts := options.NewOptions().SetMinimizeIntegerSize(false)
	//truncateTrueMinFalseOpts := options.NewOptions().
	//	SetTruncateIntegers(true).
	//	SetMinimizeIntegerSize(false)
	//truncateTrueOpts := options.NewOptions().SetTruncateIntegers(true)
	//truncateFalseOpts := options.NewOptions().SetTruncateIntegers(false)

	testCases := []struct {
		name       string
		inputfile  string
		outputfile string
		opts       *options.Options
		expectErr  bool
	}{
		//relaxed cases
		{"unified", "unified", "unified", nil, false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := jsonutil.ReadFileBytesOrPanic(fmt.Sprintf("../testdata/%s.json", tc.inputfile))
			actual, err := Convert(input, tc.opts)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			} else {
				if err != nil {
					t.Fatalf("expected nil, got error %v", err)
				}
			}
			expected := jsonutil.ReadFileBytesOrPanic(fmt.Sprintf("../testdata/%s.generated", tc.outputfile))
			if diff := cmp.Diff(string(expected), actual); diff != "" {
				t.Fatalf("Generated struct doesn't match expected, got difference %s", diff)
			}
		})
	}
}

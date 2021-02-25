package extjson

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sindbach/json-to-bson-go/jsonutil"
	"github.com/sindbach/json-to-bson-go/options"
)

func TestConvert(t *testing.T) {
	customNameOpts := options.NewOptions().SetStructName("CustomStruct")
	minimizeTrueOpts := options.NewOptions().SetMinimizeIntegerSize(true)
	minimizeFalseOpts := options.NewOptions().SetMinimizeIntegerSize(false)
	truncateTrueMinFalseOpts := options.NewOptions().
		SetTruncateIntegers(true).
		SetMinimizeIntegerSize(false)
	truncateTrueOpts := options.NewOptions().SetTruncateIntegers(true)
	truncateFalseOpts := options.NewOptions().SetTruncateIntegers(false)

	testCases := []struct {
		name       string
		inputfile  string
		outputfile string
		opts       *options.Options
		expectErr  bool
	}{
		//relaxed cases
		{"scalar values/simple", "simplejson_scalar", "simplejson_scalar", nil, false},
		{"scalar values/ext", "extjson_scalar", "simplejson_scalar", nil, false},
		{"primitive values", "extjson_primitive", "primitive", nil, false},
		{"arrays/simple", "simplejson_array", "simplejson_array", nil, false},
		{"arrays/ext", "extjson_array", "simplejson_array", nil, false},
		{"nested documents/simple", "simplejson_nested", "simplejson_nested", nil, false},
		{"nested documents/ext", "extjson_nested", "simplejson_nested", nil, false},
		{"custom struct name", "simplejson_custom_name", "simplejson_custom_name", customNameOpts, false},
		// The scalar values file contains examples of numeric values being represented as types other than float64.
		// The MinimizeIntegerSize option is true by default, so that test passes without any options as well.
		{"minimize ints true", "simplejson_scalar", "simplejson_scalar", minimizeTrueOpts, false},
		{"minimize ints false", "simplejson_minimize_false", "simplejson_minimize_false", minimizeFalseOpts, false},
		{"truncate ints true is a noop if minimize is false", "simplejson_minimize_false", "simplejson_minimize_false", truncateTrueMinFalseOpts, false},
		{"truncate true", "simplejson_minimize_false", "simplejson_truncate_true", truncateTrueOpts, false},
		{"truncate false", "simplejson_scalar", "simplejson_scalar", truncateFalseOpts, false},

		// Error cases
		{"invalid json", "simplejson_invalid", "", nil, true},
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
				return
			}
			expected := jsonutil.ReadFileBytesOrPanic(fmt.Sprintf("../testdata/%s.generated", tc.outputfile))
			if diff := cmp.Diff(string(expected), actual); diff != "" {
				t.Fatalf("Generated struct doesn't match expected, got difference %s", diff)
			}
		})
	}
}

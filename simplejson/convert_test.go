package simplejson

import (
	"fmt"
	"testing"

	"github.com/sindbach/json-to-bson-go/jsonutil"
)

func TestConvert(t *testing.T) {
	testCases := []struct {
		name      string
		filename  string
		expectErr bool
	}{
		{"scalar values", "simplejson_scalar", false},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := jsonutil.ReadFileBytesOrPanic(fmt.Sprintf("../testdata/%s.json", tc.filename))
			actual, err := Convert(input)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}

			expected := jsonutil.ReadFileBytesOrPanic(fmt.Sprintf("../testdata/%s.generated", tc.filename))
			if string(expected) != actual {
				t.Fatalf("expected generated struct %s, got %s", string(expected), actual)
			}
		})
	}
}

package options

import (
	"testing"
)

func TestConvert(t *testing.T) {
	name := "StructName"
	customNameOpts := NewOptions().SetStructName(name)
	if customNameOpts.StructName() != name {
		t.Fatalf("failed to set struct name option")
	}

	minInt := false
	minimizeTrueOpts := NewOptions().SetMinimizeIntegerSize(minInt)
	if minimizeTrueOpts.MinimizeIntegerSize() != minInt {
		t.Fatalf("failed to set minimize integer option")
	}

	truncInt := true
	truncateTrueOpts := NewOptions().SetTruncateIntegers(truncInt)
	if truncateTrueOpts.TruncateIntegers() != truncInt {
		t.Fatalf("failed to set truncate integers option")
	}
}

package axdr

import (
	"bytes"
	"testing"
)

func TestNullDataEncodesAsTagOnly(t *testing.T) {
	encoded, err := CreateAxdrNull().Encode()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(encoded, []byte{0}) {
		t.Fatalf("encoded null-data = %x, want 00", encoded)
	}
}

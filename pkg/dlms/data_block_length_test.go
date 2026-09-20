package dlms

import (
	"bytes"
	"testing"
)

func TestDataBlocksRoundTripMultiOctetLength(t *testing.T) {
	raw := bytes.Repeat([]byte{0x5a}, 300)
	getEncoded, err := CreateDataBlockGAsData(true, 7, raw).Encode()
	if err != nil {
		t.Fatal(err)
	}
	getDecoded, err := DecodeDataBlockG(&getEncoded)
	if err != nil || len(getEncoded) != 0 || !bytes.Equal(getDecoded.Result.([]byte), raw) {
		t.Fatalf("GET block remaining=%x error=%v", getEncoded, err)
	}

	setEncoded, err := CreateDataBlockSA(true, 8, raw).Encode()
	if err != nil {
		t.Fatal(err)
	}
	setDecoded, err := DecodeDataBlockSA(&setEncoded)
	if err != nil || len(setEncoded) != 0 || !bytes.Equal(setDecoded.Raw, raw) {
		t.Fatalf("SET/ACTION block remaining=%x error=%v", setEncoded, err)
	}
}

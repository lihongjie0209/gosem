package dlms

import (
	"testing"
)

func TestDecodeRLRE(t *testing.T) {
	src := decodeHexString("6300")
	rlre, err := DecodeRLRE(&src)
	if err != nil {
		t.Errorf("Failed on DecodeAARE. Err: %v", err)
	}

	if rlre.ReleaseResponseReason != nil {
		t.Errorf("Invalid ReleaseResponseReason. Should be nil but get %v", rlre.ReleaseResponseReason)
	}
}

func TestDecodeRLREWithUserInformation(t *testing.T) {
	src := decodeHexString("6328800100BE230421281F30000000097A01C161F198612BB535740660BAEDA2FC42C287E527543BBA97")
	rlre, err := DecodeRLRE(&src)
	if err != nil {
		t.Errorf("Failed on DecodeAARE. Err: %v", err)
	}

	if *rlre.ReleaseResponseReason != ReleaseResponseReasonNormal {
		t.Errorf("Invalid AssociationResult. Get %v", *rlre.ReleaseResponseReason)
	}
}

func TestDecodeRLREGuruxSecuredUserInformationLengthVariant(t *testing.T) {
	// GuruxDLMS.c reports the outer length four bytes short and the BE length
	// one byte short, while its nested octet-string length is exact. Keep this
	// captured response as a narrow interoperability regression vector.
	src := decodeHexString("6324800100BE220421")
	src = append(src, make([]byte, 33)...)
	rlre, err := DecodeRLRE(&src)
	if err != nil {
		t.Fatalf("DecodeRLRE rejected Gurux secured response: %v", err)
	}
	if rlre.ReleaseResponseReason == nil || *rlre.ReleaseResponseReason != ReleaseResponseReasonNormal {
		t.Fatalf("unexpected release response reason: %v", rlre.ReleaseResponseReason)
	}
	if len(src) != 0 {
		t.Fatalf("DecodeRLRE left %d response bytes", len(src))
	}
}

func TestDecodeRLRERejectsUnboundedLengthVariants(t *testing.T) {
	tests := []string{
		"6323800100BE220421" + string(make([]byte, 33)), // outer delta is not Gurux's four bytes
		"6324800100BE220420" + string(make([]byte, 33)), // nested length does not close the payload
		"6324800100BF220421" + string(make([]byte, 33)), // unknown trailing field
	}
	for _, encoded := range tests {
		src := []byte(encoded)
		if _, err := DecodeRLRE(&src); err == nil {
			t.Fatal("DecodeRLRE accepted malformed length variant")
		}
	}
}

package dlms

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeAARQWithoutAuthentication(t *testing.T) {
	settings, _ := NewSettingsWithoutAuthentication()
	out, err := EncodeAARQ(&settings)
	assert.NoError(t, err)

	expected := decodeHexString("601DA109060760857405080101BE10040E01000000065F1F040000181F0100")
	assert.Equal(t, expected, out)
}

func TestEncodeAARQWithLowAuthentication(t *testing.T) {
	settings, _ := NewSettingsWithLowAuthentication([]byte("12345678"))
	out, err := EncodeAARQ(&settings)
	assert.NoError(t, err)

	expected := decodeHexString("6036A1090607608574050801018A0207808B0760857405080201AC0A80083132333435363738BE10040E01000000065F1F040000181F0100")
	assert.Equal(t, expected, out)

	settings.Password = nil
	_, err = EncodeAARQ(&settings)
	assert.Error(t, err)
}

func TestEncodeAARQWithLowAuthenticationAndCipher(t *testing.T) {
	ciphering, _ := NewCiphering(
		SecurityLevelDedicatedKey,
		SecurityEncryption|SecurityAuthentication,
		decodeHexString("4349520000000001"),
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
		0x00000107,
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
	)
	ciphering.DedicatedKey = decodeHexString("E803739DBE338C3A790D8D1B12C63FE2")

	settings, _ := NewSettingsWithLowAuthenticationAndCiphering([]byte("JuS66BCZ"), ciphering)
	settings.MaxPduRecvSize = 512

	out, err := EncodeAARQ(&settings)
	assert.NoError(t, err)

	expected := decodeHexString("6066A109060760857405080103A60A040843495200000000018A0207808B0760857405080201AC0A80084A7553363642435ABE340432213030000001078E6341442275404C816C6BED3E33AE809EC51E1D0E428BE8F5F643E26C3DD89FD2E3F2220097124F58E0F4")
	assert.Equal(t, expected, out)

	assert.Equal(t, uint32(0x00000108), settings.Ciphering.UnicastKeyIC)

	settings.Ciphering.UnicastKey = nil
	_, err = EncodeAARQ(&settings)
	assert.Error(t, err)
}

func TestEncodeAARQReservesInvocationCounterBeforeCiphering(t *testing.T) {
	ciphering, _ := NewCiphering(
		SecurityLevelGlobalKey,
		SecurityEncryption|SecurityAuthentication,
		decodeHexString("4349520000000001"),
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
		7,
		decodeHexString("00112233445566778899AABBCCDDEEFF"),
	)
	wantErr := errors.New("persistence unavailable")
	var level SecurityLevel
	var counter uint32
	ciphering.BeforeInvocationCounter = func(gotLevel SecurityLevel, gotCounter uint32) error {
		level, counter = gotLevel, gotCounter
		return wantErr
	}
	settings, _ := NewSettingsWithLowAuthenticationAndCiphering([]byte("secret"), ciphering)

	_, err := EncodeAARQ(&settings)
	assert.ErrorIs(t, err, wantErr)
	assert.Equal(t, SecurityLevelGlobalKey, level)
	assert.Equal(t, uint32(7), counter)
	assert.Equal(t, uint32(7), settings.Ciphering.UnicastKeyIC)
}

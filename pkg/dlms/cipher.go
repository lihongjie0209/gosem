package dlms

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"fmt"

	"gitlab.com/circutor-library/gosem/pkg/axdr"
)

type Cipher struct {
	Tag          CosemTag
	Security     Security
	SystemTitle  []byte
	Key          []byte
	AuthKey      []byte
	FrameCounter uint32
}

func CipherData(cfg Cipher, data []byte) ([]byte, error) {
	// Generate a new AES cipher using our 32 byte long key
	c, err := aes.NewCipher(cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// GCM or Galois/Counter Mode, is a mode of operation for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCMWithTagSize(c, 12)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// Initialization vector (or Nonce)
	iv := make([]byte, gcm.NonceSize())
	copy(iv, cfg.SystemTitle)
	binary.BigEndian.PutUint32(iv[8:], cfg.FrameCounter)

	// Associated data
	ad := make([]byte, 17)
	ad[0] = byte(cfg.Security)
	copy(ad[1:], cfg.AuthKey)

	// Ciphered data prefix
	size, _ := axdr.EncodeLength(5 + len(data) + 12)
	dst := make([]byte, 6+len(size))
	dst[0] = byte(cfg.Tag)
	copy(dst[1:], size)
	dst[1+len(size)] = byte(cfg.Security)
	binary.BigEndian.PutUint32(dst[2+len(size):], cfg.FrameCounter)

	// Encrypt data
	return gcm.Seal(dst, iv, data, ad), nil
}

func DecipherData(cfg *Cipher, data []byte) ([]byte, error) {
	// Check COSEM tag
	if data[0] != byte(cfg.Tag) {
		return nil, ErrWrongTag(0, data[0], byte(cfg.Tag))
	}
	data = data[1:]

	// Check length
	_, length, err := axdr.DecodeLength(&data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode length: %w", err)
	}

	if len(data) != int(length) {
		err = ErrWrongLength(int(length), len(data))
		return nil, err
	}

	// Check security level
	if data[0] != byte(cfg.Security) {
		return nil, errors.New("wrong security level")
	}
	data = data[1:]

	// Generate a new AES cipher using our 32 byte long key
	c, err := aes.NewCipher(cfg.Key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// GCM or Galois/Counter Mode, is a mode of operation for symmetric key cryptographic block ciphers
	// - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	gcm, err := cipher.NewGCMWithTagSize(c, 12)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	// Initialization vector (or Nonce)
	iv := make([]byte, gcm.NonceSize())
	copy(iv, cfg.SystemTitle)
	copy(iv[8:], data[:4])
	data = data[4:]

	// Save frame counter
	cfg.FrameCounter = binary.BigEndian.Uint32(iv[8:])

	// Associated data
	ad := make([]byte, 17)
	ad[0] = byte(cfg.Security)
	copy(ad[1:], cfg.AuthKey)

	// Decrypt data
	data, err = gcm.Open(nil, iv, data, ad)
	if err != nil {
		return nil, fmt.Errorf("failed to decipher data: %w", err)
	}

	return data, nil
}

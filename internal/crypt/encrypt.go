package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
)

const (
	nonceSize  = 12
	aesKeySize = 32
)

func EncryptHybrid(pub *rsa.PublicKey, plain []byte) ([]byte, error) {
	if pub == nil {
		return nil, errors.New("public key is nil")
	}

	aesKey := make([]byte, aesKeySize)
	if _, err := io.ReadFull(crand.Reader, aesKey); err != nil {
		return nil, fmt.Errorf("read aes key: %w", err)
	}

	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(crand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("read nonce: %w", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("aes.NewCipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("cipher.NewGCM: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plain, nil)

	encKey, err := rsa.EncryptOAEP(sha256.New(), crand.Reader, pub, aesKey, nil)
	if err != nil {
		return nil, fmt.Errorf("rsa.EncryptOAEP: %w", err)
	}

	out := make([]byte, 0, len(encKey)+len(nonce)+len(ciphertext))
	out = append(out, encKey...)
	out = append(out, nonce...)
	out = append(out, ciphertext...)

	return out, nil
}

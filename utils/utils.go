package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

const (
	NONCE_SIZE = 12
	MAX_KEY_SIZE = 20
)

func EncryptAES256(key []byte, plaintext string) (string, error) {
	// CHECK KEY SIZE AND SIZE OF PLAINTEXT
	// if key size < 32 there is no aes 256 ecryption
	// so we need generate "nonce" for fill all bytes of key

	if len(key) > MAX_KEY_SIZE {
		// cut key to 20 bytes
		key = key[:MAX_KEY_SIZE]
	}

	if len(plaintext) < 16 {
		return "", errors.New("plaintext size less then 16")
	}

	bplaintext := []byte(plaintext)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", nil
	}

	nonce := make([]byte, NONCE_SIZE)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, bplaintext, nil)

	ciphertextBase64 := base64.StdEncoding.EncodeToString(ciphertext)

	return ciphertextBase64, nil
}

func DecryptAES256(key []byte, base64ciphertext string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(base64ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("nonceSize is bigger than encrypted text")
	}

	// split the nonce from the ciptertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
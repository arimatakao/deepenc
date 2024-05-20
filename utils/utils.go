package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

const (
	EMPTY_SYMBOLS = "                "
)

func EncryptAES256(key []byte, plaintext string) (string, error) {
	if len(key) == 0 {
		return "", errors.New("key is empty")
	}
	if len(plaintext) == 0 {
		return "", errors.New("plaintext is empty")
	}

	keyHashed := sha256.Sum256(key)

	if len(plaintext) < 16 {
		plaintext += EMPTY_SYMBOLS[len(plaintext):]
	}

	bplaintext := []byte(plaintext)

	block, err := aes.NewCipher(keyHashed[:])
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", nil
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesgcm.Seal(nonce, nonce, bplaintext, nil)

	ciphertextBase64 := base64.StdEncoding.EncodeToString(ciphertext)

	return ciphertextBase64, nil
}

func DecryptAES256(key []byte, base64ciphertext string) (string, error) {
	if len(key) == 0 {
		return "", errors.New("key is empty")
	}
	if len(base64ciphertext) == 0 {
		return "", errors.New("base64ciphertext is empty")
	}
	keyHashed := sha256.Sum256(key)

	ciphertext, err := base64.StdEncoding.DecodeString(base64ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(keyHashed[:])
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

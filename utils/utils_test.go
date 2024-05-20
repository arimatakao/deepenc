package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestCaseEncryption struct {
	Name      string
	Key       []byte
	Plaintext string
	WithError bool
}

func TestEncryptAES256(t *testing.T) {
	key := "testkey"
	plaintext := "plaintext1234567"
	plaintextShort := "plaintext"

	cases := []TestCaseEncryption{
		{
			Name:      "key is empty",
			Key:       []byte{},
			Plaintext: plaintext,
			WithError: true,
		},
		{
			Name:      "plaintext is empty",
			Key:       []byte(key),
			Plaintext: "",
			WithError: true,
		},
		{
			Name:      "key is fine and plaintext is short",
			Key:       []byte("testkey"),
			Plaintext: plaintextShort,
			WithError: false,
		},
		{
			Name:      "key and plaintext is fine",
			Key:       []byte("testkey"),
			Plaintext: plaintext,
			WithError: false,
		},
	}

	for _, testCase := range cases {
		result, err := EncryptAES256(testCase.Key, testCase.Plaintext)

		if testCase.WithError {
			assert.Empty(t, result, testCase.Name)
			assert.NotNil(t, err, testCase.Name)
		} else {
			assert.NotEmpty(t, result, testCase.Name)
			assert.Nil(t, err, testCase.Name)
		}
	}
}

type TestCaseDecryption struct {
	Name             string
	Key              []byte
	CiphertextBase64 string
	ExpectedResult   string
	WithError        bool
}

func TestDecryptAES256(t *testing.T) {
	key := "testkey"

	cases := []TestCaseDecryption{
		{
			Name:             "key is empty",
			Key:              []byte{},
			CiphertextBase64: "123",
			ExpectedResult:   "",
			WithError:        true,
		},
		{
			Name:             "base64ciphertext is empty",
			Key:              []byte(key),
			CiphertextBase64: "",
			ExpectedResult:   "",
			WithError:        true,
		},
	}

	for _, testCase := range cases {
		result, err := DecryptAES256(testCase.Key, testCase.CiphertextBase64)

		if testCase.WithError {
			assert.Empty(t, result, testCase.Name)
			assert.NotNil(t, err, testCase.Name)
		} else {
			assert.NotEmpty(t, result, testCase.Name)
			assert.Nil(t, err, testCase.Name)
		}
	}
}

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateHash(t *testing.T) {
	data := []byte("test data")
	key := "secret"

	hash := CreateHash(data, key)
	assert.NotEmpty(t, hash)

	// Same data and key should produce same hash
	hash2 := CreateHash(data, key)
	assert.Equal(t, hash, hash2)

	// Different key should produce different hash
	hash3 := CreateHash(data, "different_key")
	assert.NotEqual(t, hash, hash3)

	// Different data should produce different hash
	hash4 := CreateHash([]byte("different data"), key)
	assert.NotEqual(t, hash, hash4)
}

func TestValidateHash(t *testing.T) {
	data := []byte("test data")
	key := "secret"

	hash := CreateHash(data, key)

	// Valid hash should return nil (no error)
	err := ValidateHash(data, key, hash)
	assert.NoError(t, err)

	// Invalid hash should return error
	err = ValidateHash(data, key, "invalid_hash")
	assert.Error(t, err)

	// Wrong key should return error
	err = ValidateHash(data, "wrong_key", hash)
	assert.Error(t, err)

	// Different data should return error
	err = ValidateHash([]byte("different data"), key, hash)
	assert.Error(t, err)
}

func TestCreateHashEmptyData(t *testing.T) {
	data := []byte{}
	key := "secret"

	hash := CreateHash(data, key)
	assert.NotEmpty(t, hash)
}

func TestCreateHashEmptyKey(t *testing.T) {
	data := []byte("test data")
	key := ""

	hash := CreateHash(data, key)
	assert.NotEmpty(t, hash)
}

func TestValidateHashInvalidHex(t *testing.T) {
	data := []byte("test data")
	key := "secret"

	// Invalid hex string should return error
	err := ValidateHash(data, key, "not-a-hex-string")
	assert.Error(t, err)
}

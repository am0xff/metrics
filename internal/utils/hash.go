package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

func CreateHash(data []byte, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

func ValidateHash(data []byte, key, headerHash string) error {
	sig, err := hex.DecodeString(headerHash)
	if err != nil {
		return err
	}
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write(data)
	expected := mac.Sum(nil)
	if !hmac.Equal(expected, sig) {
		return errors.New("hash mismatch")
	}
	return nil
}

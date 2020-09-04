package utils

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"os"

	"github.com/juju/errors"
)

// NewContext returns default context
func NewContext() context.Context {
	return context.Background()
}

// GetHashString returns sha256 hash sum with hex encoding of bytes
func GetHashString(bytes []byte) (string, error) {
	h := sha1.New()
	_, err := h.Write(bytes)
	if err != nil {
		return "", errors.Trace(err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// FileExists checks if file exists
func FileExists(path string) (exist bool, err error) {
	if _, err = os.Stat(path); err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

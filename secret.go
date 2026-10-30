package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

// Provider API keys are encrypted at rest with AES-GCM when VSCHAT_SECRET is set.
// Without it they are stored as plain text, as before.
const encPrefix = "enc:v1:"

type sealer struct{ aead cipher.AEAD }

func newSealer(secret string) *sealer {
	if secret == "" {
		return nil
	}
	key := sha256.Sum256([]byte("vschat-key-encryption:" + secret))
	block, _ := aes.NewCipher(key[:])
	aead, _ := cipher.NewGCM(block)
	return &sealer{aead: aead}
}

func (s *sealer) seal(plain string) string {
	if s == nil || plain == "" || strings.HasPrefix(plain, encPrefix) {
		return plain
	}
	nonce := make([]byte, s.aead.NonceSize())
	rand.Read(nonce)
	return encPrefix + base64.RawStdEncoding.EncodeToString(s.aead.Seal(nonce, nonce, []byte(plain), nil))
}

var errNoSecret = errors.New("API key is encrypted but VSCHAT_SECRET is not set (or changed)")

func (s *sealer) open(stored string) (string, error) {
	if !strings.HasPrefix(stored, encPrefix) {
		return stored, nil
	}
	if s == nil {
		return "", errNoSecret
	}
	data, err := base64.RawStdEncoding.DecodeString(stored[len(encPrefix):])
	n := s.aead.NonceSize()
	if err != nil || len(data) < n {
		return "", errNoSecret
	}
	plain, err := s.aead.Open(nil, data[:n], data[n:], nil)
	if err != nil {
		return "", errNoSecret
	}
	return string(plain), nil
}

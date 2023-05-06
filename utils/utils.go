package utils

import (
	"crypto/sha256"
	"encoding/base64"
)

func Hash(s string) string {
	return "" //TODO
}

func calcAddress(key string) string {
	hash := sha256.Sum256([]byte(key))
	addr := base64.StdEncoding.EncodeToString(hash[:])
	return addr
}

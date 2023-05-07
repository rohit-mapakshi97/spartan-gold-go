package utils

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
)

const (
	HASH_ALG       = crypto.SHA256
	SIG_ALG        = crypto.SHA256 // there is no equivalent RSA-SHA256 in Go
	MODULUS_LENGTH = 512
)

//	func Hash(s string) string {
//		return "" //TODO
//	}
func Hash(s string) []byte {
	hash := crypto.Hash(HASH_ALG).New()
	hash.Write([]byte(s))
	return hash.Sum(nil)
}

func calcAddress(key string) string {
	hash := sha256.Sum256([]byte(key))
	addr := base64.StdEncoding.EncodeToString(hash[:])
	return addr
}

func Sign(privKey []byte, msg interface{}) (string, error) {
	signer := rsa.PrivateKey{}
	if err := json.Unmarshal(privKey, &signer); err != nil {
		return "", err
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	hashed := sha256.Sum256(msgBytes)
	signature, err := signer.Sign(rand.Reader, hashed[:], SIG_ALG)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(signature), nil
}

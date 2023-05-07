package blockchain

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

type Client struct {
	name       string
	address    string
	publicKey  *pem.Block
	privateKey *pem.Block
}

func NewClient(options map[string]interface{}) *Client {
	c := &Client{}

	if options["name"] != nil {
		c.name = options["name"].(string)
	}
	// if options["publicKey"] == nil && options["privateKey"] == nil {
	reader := rand.Reader
	bitSize := 512
	var err error
	// var publicKey rsa.PublicKey
	key, err := rsa.GenerateKey(reader, bitSize)

	_ = err
	var privateKeyBytes []byte = x509.MarshalPKCS1PrivateKey(key)
	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}
	c.privateKey = privateKeyBlock

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	publicKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}
	c.publicKey = publicKeyBlock
	// c.address = calcAddress(c.publicKey)
	// } else {
	// 	c.publicKey = options["publicKey"]
	// 	c.privateKey = options["privateKey"]
	// }
	c.address = calcAddress(c.name) // change to public key later
	// fmt.Printf("PUB KEY BLOCK: %T\n", publicKeyBlock)
	return c
}

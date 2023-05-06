package blockchain

type Fakenet struct{}
package blockchain

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
)

type FakeNet struct {
	clients map[string]interface{}
}

func (fn *FakeNet) Register(clientList ...*Client) {
	for _, client := range clientList {
		fn.clients[client.address] = client
	}
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

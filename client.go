package blockchain

import (
	"encoding/pem"
)

type Client struct {
	name       string
	address    string
	publicKey  *pem.Block
	privateKey *pem.Block
}

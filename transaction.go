package blockchain

import (
	"encoding/json"
	// Try import "./utils/utils/go"
)

const TX_CONST = "TX"

type Output struct {
	amount  int    `json:"amount"`
	address string `json:"address"`
}

type Transaction struct {
	from    string                 `json:"from"`
	nonce   int                    `json:"nonce"`
	pubKey  string                 `json:"pubKey"`
	sig     string                 `json:"sig,omitempty"`
	fee     int                    `json:"fee"`
	outputs []Output               `json:"outputs"`
	data    map[string]interface{} `json:"data"`
}

func NewTransaction(from string, nonce int, pubKey string, sig string, outputs []Output, fee int, data map[string]interface{}) *Transaction {
	t := &Transaction{
		from:    from,
		nonce:   nonce,
		pubKey:  pubKey,
		sig:     sig,
		fee:     fee,
		outputs: outputs,
		data:    data,
	}
	return t
}

func (t *Transaction) ID() string {
	txData := map[string]interface{}{
		"from":    t.from,
		"nonce":   t.nonce,
		"pubKey":  t.pubKey,
		"outputs": t.outputs,
		"fee":     t.fee,
		"data":    t.data,
	}
	txDataBytes, _ := json.Marshal(txData)
	return utils.hash(TX_CONST + string(txDataBytes))
}

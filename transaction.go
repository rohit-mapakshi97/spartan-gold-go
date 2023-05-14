package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

/**
 * A transaction comes from a single account, specified by "address". For
 * each account, transactions have an order established by the nonce. A
 * transaction should not be accepted if the nonce has already been used.
 * (Nonces are in increasing order, so it is easy to determine when a nonce
 * has been used.)
 */

type Output struct {
	Address string
	Amount  uint32
}

type TransactionInfo struct {
	From    string
	Nonce   uint32
	Pubkey  rsa.PublicKey
	Fee     uint32
	Outputs []Output
	Data    []byte
}

type Transaction struct {
	Info TransactionInfo
	Sig  []byte
}

func NewTransaction(from string, nonce uint32, pubkey *rsa.PublicKey, sig []byte, fee uint32, outputs []Output, data []byte) *Transaction {
	var tx Transaction
	tx.Info.From = from
	tx.Info.Nonce = nonce
	tx.Info.Pubkey = *pubkey
	tx.Info.Fee = fee
	if len(outputs) == 0 {
		panic("outputs is empty")
	}
	tx.Info.Outputs = make([]Output, len(outputs))
	copy(tx.Info.Outputs, outputs)
	tx.Info.Data = make([]byte, len(data))
	copy(tx.Info.Data, data)

	tx.Sig = make([]byte, len(sig))
	copy(tx.Sig, sig)
	return &tx
}

/**
 * A transaction's ID is derived from its contents.
 */
func (tx *Transaction) Id() string {
	return tx.GetHashStr()
}

/**
 * Signs a transaction and stores the signature in the transaction.
 */
func (tx *Transaction) Sign(privKey *rsa.PrivateKey) []byte {
	rng := rand.Reader
	hashed := (&(*tx).Info).GetHash()
	signature, err := rsa.SignPKCS1v15(rng, privKey, crypto.SHA256, hashed)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from signing: %s\n", err)
		return nil
	}
	tx.Sig = make([]byte, len(signature))
	copy(tx.Sig, signature)
	return signature
}

/**
 * Verifies that there is currently sufficient gold for the transaction.
 */

func (tx *Transaction) ValidSignature() bool {
	hashed := (&(*tx).Info).GetHash()
	err := rsa.VerifyPKCS1v15(&tx.Info.Pubkey, crypto.SHA256, hashed[:], (*tx).Sig)
	return err == nil
}

func TransactionToBytes(tx *Transaction) []byte {
	data, err := json.Marshal(tx)
	if err != nil {
		return nil
	}
	return data
}

func BytesToTransaction(data []byte) *Transaction {
	var tx Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		return nil
	}
	return &tx
}

func TransactionInfoToBytes(tx *TransactionInfo) []byte {
	data, err := json.Marshal(tx)
	if err != nil {
		return nil
	}
	return data
}

func BytesToTransactionInfo(data []byte) *TransactionInfo {
	var txInfo TransactionInfo
	if err := json.Unmarshal(data, &txInfo); err != nil {
		return nil
	}
	return &txInfo
}

func (tran *Transaction) ToString() string {
	info := fmt.Sprintf("from : %s\n"+
		"nonce: %d\n"+
		"pubkey: \n\tN: %x\n\tE: %x\n"+
		"fee: %d\n", (*tran).Info.From, (*tran).Info.Nonce,
		(*tran).Info.Pubkey.N, (*tran).Info.Pubkey.E, (*tran).Info.Fee)

	outputs := "outputs: [\n"
	for _, v := range (*tran).Info.Outputs {
		outputs = outputs + fmt.Sprintf("\t{address: %s amount: %d}\n", v.Address, v.Amount)
	}
	outputs = outputs + "]\n"
	info = info + outputs
	data := fmt.Sprintf("data: %s\nsig:%x", hex.EncodeToString(tran.Info.Data), hex.EncodeToString((*tran).Sig))
	info = info + data

	return info
}

func (txInfo *TransactionInfo) GetHash() []byte {
	data := TransactionInfoToBytes(txInfo)
	tx_hash := sha256.Sum256(data)
	return tx_hash[:]
}

func (tx *Transaction) GetHashStr() string {
	data := TransactionToBytes(tx)
	hashed := sha256.Sum256(data)
	return hex.EncodeToString(hashed[:])
}

func (tx *Transaction) TotalOutput() uint32 {
	var amount uint32 = 0
	for _, v := range (*tx).Info.Outputs {
		amount += v.Amount
	}
	amount += (*tx).Info.Fee
	return amount
}

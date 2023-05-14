package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

type TransactionType struct {
	Id string
	Tx Transaction
}

type NextNonceType struct {
	Id    string
	Nonce uint32
}

type BalanceType struct {
	Id      string
	Balance uint32
}

type Block struct {
	PrevBlockHash  string
	Target         big.Int
	Proof          uint32
	Balances       []BalanceType
	NextNonce      []NextNonceType
	Transactions   []TransactionType
	ChainLength    uint32
	Timestamp      time.Time
	RewardAddr     string
	CoinbaseReward uint32
}

func (block *Block) FindTransactionIndex(id string) int {
	index := int(-1)
	for i, v := range block.Transactions {
		if v.Id == id {
			index = i
			break
		}
	}
	return index
}

func (block *Block) FindNextNonceIndex(id string) int {
	index := int(-1)
	for i, v := range block.NextNonce {
		if v.Id == id {
			index = i
			break
		}
	}
	return index
}

func (block *Block) FindBalanceIndex(id string) int {
	index := int(-1)
	for i, v := range block.Balances {
		if v.Id == id {
			index = i
			break
		}
	}
	return index
}

func NewBlock(rewardAddr string, prevBlock *Block, target *big.Int, coinbaseReward uint32) *Block {
	var block Block
	block.Target = *target
	block.Proof = 0

	if prevBlock != nil {
		hashHexStr := prevBlock.GetHash()
		block.PrevBlockHash = hashHexStr
	}

	// Get the balances and nonces from the previous block, if available.
	block.Balances = make([]BalanceType, 0)
	if prevBlock != nil && (*prevBlock).Balances != nil {
		block.Balances = append(block.Balances, (*prevBlock).Balances...)
	}

	block.NextNonce = make([]NextNonceType, 0)
	if prevBlock != nil && (*prevBlock).NextNonce != nil {
		block.NextNonce = append(block.NextNonce, (*prevBlock).NextNonce...)
	}

	// Storing transactions
	block.Transactions = make([]TransactionType, 0)

	// Used to determine the winner between competing chains.
	// Note that this is a little simplistic -- an attacker
	// could make a long, but low-work chain.  However, this works
	// well enough for us.
	block.ChainLength = 0
	if prevBlock != nil {
		block.ChainLength = (*prevBlock).ChainLength + 1
	}

	block.Timestamp = time.Now()

	// The address that will gain both the coinbase reward and transaction fees,
	// assuming that the block is accepted by the network.
	block.RewardAddr = rewardAddr
	block.CoinbaseReward = coinbaseReward
	return &block
}

func BlockToBytes(block *Block) []byte {
	data, err := json.Marshal(block)
	if err != nil {
		return nil
	}
	return data
}

func BytesToBlock(data []byte) *Block {
	var block Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil
	}
	return &block
}

func (block *Block) GetHash() string {
	block4hash := *block
	block4hash.Balances = nil
	block4hash.NextNonce = nil

	blockData := BlockToBytes(&block4hash)
	var blockHash [32]byte

	blockHash = sha256.Sum256(blockData)
	return hex.EncodeToString(blockHash[:])

}

func (block *Block) GetHashStr() string {
	block4hash := *block
	block4hash.Balances = nil
	block4hash.NextNonce = nil

	blockData := BlockToBytes(&block4hash)
	var blockHash [32]byte

	blockHash = sha256.Sum256(blockData)
	return hex.EncodeToString(blockHash[:])

}

func (block *Block) IsGenesisBlock() bool {
	return block.ChainLength == 0
}

func (block *Block) hasValidProof() bool {
	block4hash := *block
	block4hash.Balances = nil
	block4hash.NextNonce = nil
	data := BlockToBytes(&block4hash)
	block_hash := sha256.Sum256(data)
	block_value := big.NewInt(0)
	block_value.SetBytes(block_hash[:])

	return block_value.Cmp(&(*block).Target) < 0
}

func (block *Block) AddTransaction(tx *Transaction) bool {
	if (*block).Contains(tx) {
		fmt.Printf("Duplicate transaction %s", tx.Id())
		return false
	} else if (*tx).Sig == nil {
		fmt.Printf("Unsigned transaction %s", tx.Id())
		return false
	} else if !tx.ValidSignature() {
		fmt.Printf("Invalid signature for transaction %s", tx.Id())
		return false
	} else if !block.HasSufficientFund(tx) {
		fmt.Printf("Insufficient fund for transaction %s", tx.Id())
		return false
	}

	nextNonceIndex := block.FindNextNonceIndex((*tx).Info.From)
	if nextNonceIndex == -1 {
		addNextNonce := NextNonceType{Id: (*tx).Info.From, Nonce: 0}
		(*block).NextNonce = append((*block).NextNonce, addNextNonce)
		nextNonceIndex = len((*block).NextNonce) - 1
	}
	expectedNonce := (*block).NextNonce[nextNonceIndex].Nonce

	if expectedNonce > (*tx).Info.Nonce {
		fmt.Printf("Replayed transaction %s", tx.Id())
		return false
	} else if expectedNonce < (*tx).Info.Nonce {
		fmt.Printf("Out of order transaction %s", tx.Id())
		return false
	}
	var nextNonce uint32 = expectedNonce + 1
	(*block).NextNonce[nextNonceIndex].Nonce = nextNonce

	var txId string = tx.Id()
	txData := TransactionType{Id: txId, Tx: *tx}
	(*block).Transactions = append((*block).Transactions, txData)

	var senderBalance uint32 = block.BalanceOf((*tx).Info.From)
	senderBalanceIndex := block.FindBalanceIndex((*tx).Info.From)
	(*block).Balances[senderBalanceIndex].Balance = senderBalance - tx.TotalOutput()

	for _, output := range (*tx).Info.Outputs {
		var oldBalance uint32 = (*block).BalanceOf(output.Address)
		oldBalanceId := block.FindBalanceIndex(output.Address)
		if oldBalanceId == -1 {
			balanceAcct := BalanceType{Id: output.Address, Balance: oldBalance + output.Amount}
			(*block).Balances = append((*block).Balances, balanceAcct)
		} else {
			(*block).Balances[oldBalanceId].Balance = oldBalance + output.Amount
		}
	}

	return true
}

func (block *Block) Rerun(prevBlock *Block) bool {

	if prevBlock == nil {
		return false
	}

	// Setting balances to the previous block's balances.
	block.Balances = make([]BalanceType, 0)
	if prevBlock != nil && (*prevBlock).Balances != nil {
		block.Balances = append(block.Balances, (*prevBlock).Balances...)
	}

	// copy NextNounce from previous block
	block.NextNonce = make([]NextNonceType, 0)
	if prevBlock != nil && (*prevBlock).NextNonce != nil {
		block.NextNonce = append(block.NextNonce, (*prevBlock).NextNonce...)
	}

	// Adding coinbase reward for prevBlock.
	if (*prevBlock).RewardAddr != "" {
		var winnerBalance uint32 = (*prevBlock).BalanceOf((*prevBlock).RewardAddr)
		index := block.FindBalanceIndex((*prevBlock).RewardAddr)

		if index == -1 {
			newBalance := BalanceType{Id: (*prevBlock).RewardAddr, Balance: winnerBalance + prevBlock.CoinbaseReward}
			(*block).Balances = append((*block).Balances, newBalance)
		} else {
			(*block).Balances[index].Balance = winnerBalance + prevBlock.CoinbaseReward
		}
	}

	// Re-enter all transactions
	txMap := make([]TransactionType, len((*block).Transactions))
	copy(txMap, (*block).Transactions)
	(*block).Transactions = make([]TransactionType, 0)
	for _, v := range txMap {
		if !block.AddTransaction(&v.Tx) {
			return false
		}
	}
	return true
}

// Gets the available gold of a user identified by an address.
func (block *Block) BalanceOf(address string) uint32 {
	index := block.FindBalanceIndex(address)
	if index > -1 {
		return block.Balances[index].Balance
	} else {
		return 0
	}
}

func (block *Block) HasSufficientFund(tx *Transaction) bool {
	var totalOutput uint32 = (*tx).TotalOutput()
	return totalOutput <= (*block).BalanceOf(tx.Info.From)
}

/**
 * The total amount of gold paid to the miner who produced this block,
 * if the block is accepted.  This includes both the coinbase transaction
 * and any transaction fees.
 */
func (block *Block) TotalRewards() uint32 {
	var total uint32 = 0
	for _, v := range (*block).Transactions {
		total += v.Tx.Info.Fee
	}
	total += (*block).CoinbaseReward
	return total
}

/**
 * Determines whether a transaction is in the block.  Note that only the
 * block itself is checked; if it returns false, the transaction might
 * still be included in one of its ancestor blocks.*/
func (block *Block) Contains(tx *Transaction) bool {
	index := block.FindTransactionIndex(tx.Id())
	return index > -1
}

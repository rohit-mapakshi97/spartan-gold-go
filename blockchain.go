package main

import (
	"spartangold/utils"
)

// Network message constants
const MISSING_BLOCK string = "MISSING_BLOCK"
const POST_TRANSACTION string = "POST_TRANSACTION"
const PROOF_FOUND string = "PROOF_FOUND"
const START_MINING string = "START_MINING"

// Constants for mining
const NUM_ROUNDS_MINING uint32 = 2000

// Constants related to proof-of-work target
const POW_BASE_TARGET_STR string = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
const POW_LEADING_ZEROES uint32 = 15

// Constants for mining rewards and default transaction fees
const COINBASE_AMT_ALLOWED uint32 = 25
const DEFAULT_TX_FEE uint32 = 1

// If a block is 6 blocks older than the current block, it is considered
// confirmed, for no better reason than that is what Bitcoin does.
// Note that the genesis block is always considered to be confirmed.
const CONFIRMED_DEPTH uint32 = 6

// Produces a new genesis block, giving the specified client balances
func MakeGenesisDefault(starting_balances map[string]uint32) *Block {
	if starting_balances == nil {
		panic("makeGenesis(...): starting_balances cannot be nil")
	}

	target := utils.CalcTarget(POW_LEADING_ZEROES, POW_BASE_TARGET_STR)
	genesis := NewBlock("", nil, target, COINBASE_AMT_ALLOWED)

	for client_address, client_balance := range starting_balances {
		newBalance := BalanceType{Id: client_address, Balance: client_balance}
		(*genesis).Balances = append((*genesis).Balances, newBalance)
	}

	return genesis
}

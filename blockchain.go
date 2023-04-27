package blockchain

import (
	"math/big"
	"reflect"
)

// Network message constants
const MISSING_BLOCK = "MISSING_BLOCK"
const POST_TRANSACTION = "POST_TRANSACTION"
const PROOF_FOUND = "PROOF_FOUND"
const START_MINING = "START_MINING"

// Constants for mining
const NUM_ROUNDS_MINING = 2000

// Constants related to proof-of-work target
const POW_BASE_TARGET_STR = "0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"

var POW_BASE_TARGET, _ = new(big.Int).SetString(POW_BASE_TARGET_STR, 0)

const POW_LEADING_ZEROES = 15

// Constants for mining rewards and default transaction fees
const COINBASE_AMT_ALLOWED = 25
const DEFAULT_TX_FEE = 1

// If a block is 6 blocks older than the current block, it is considered
// confirmed, for no better reason than that is what Bitcoin does.
// Note that the genesis block is always considered to be confirmed.
const CONFIRMED_DEPTH = 6

type Blockchain struct {
	cfg CFG
}

/**
   * @param {Object} cfg - Settings for the blockchain.
   * @param {Class} cfg.blockClass - Implementation of the Block class.
   * @param {Class} cfg.transactionClass - Implementation of the Transaction class.
   * @param {Map} [cfg.clientBalanceMap] - Mapping of clients to their starting balances.
   * @param {Object} [cfg.startingBalances] - Mapping of client addresses to their starting balances.
   * @param {number} [cfg.powLeadingZeroes] - Number of leading zeroes required for a valid proof-of-work.
   * @param {number} [cfg.coinbaseAmount] - Amount of gold awarded to a miner for creating a block.
   * @param {number} [cfg.defaultTxFee] - Amount of gold awarded to a miner for accepting a transaction,
   *    if not overridden by the client.
   * @param {number} [cfg.confirmedDepth] - Number of blocks required after a block before it is
   *    considered confirmed.
**/
type CFG struct {
	blockClass       reflect.Type //[DONE] confirm if this is a object or a class
	transactionClass reflect.Type
	clientBalanceMap map[Client]int //Mapping of clients to their starting balances. the type colud be differnt here [TODO]
	startingBalances map[string]int
	powLeadingZeroes int
	coinbaseAmount   int
	defaultTxFee     int
	confirmedDepth   int
	powTarget        *big.Int
}

func (bc *Blockchain) MISSING_BLOCK() string {
	return MISSING_BLOCK
}

func (bc *Blockchain) POST_TRANSACTION() string {
	return POST_TRANSACTION
}

func (bc *Blockchain) PROOF_FOUND() string {
	return PROOF_FOUND
}

func (bc *Blockchain) START_MINING() string {
	return START_MINING
}

func (bc *Blockchain) NUM_ROUNDS_MINING() int {
	return NUM_ROUNDS_MINING
}

// Configurable properties.
func (bc *Blockchain) POW_TARGET() *big.Int {
	return bc.cfg.powTarget
}

func (bc *Blockchain) COINBASE_AMT_ALLOWED() int {
	return bc.cfg.coinbaseAmount
}

func (bc *Blockchain) DEFAULT_TX_FEE() int {
	return bc.cfg.defaultTxFee
}

func (bc *Blockchain) CONFIRMED_DEPTH() int {
	return bc.cfg.confirmedDepth
}

func (bc *Blockchain) makeGenesis(blockClass reflect.Type, transactionClass reflect.Type, powLeadingZeroes int, coinbaseAmount int, defaultTxFee int, confirmedDepth int, clientBalanceMap map[Client]int, startingBalances map[string]int) Block {
	if clientBalanceMap != nil && startingBalances != nil {
		panic("You may set clientBalanceMap OR set startingBalances, but not both.")
	}

	// Setting blockchain configuration
	var cfg CFG
	cfg.blockClass = blockClass
	cfg.transactionClass = transactionClass
	cfg.powLeadingZeroes = powLeadingZeroes
	cfg.coinbaseAmount = COINBASE_AMT_ALLOWED
	cfg.defaultTxFee = defaultTxFee
	cfg.confirmedDepth = confirmedDepth
	cfg.clientBalanceMap = clientBalanceMap
	cfg.startingBalances = startingBalances

	cfg.powTarget = new(big.Int).Rsh(POW_BASE_TARGET, uint(cfg.powLeadingZeroes))

	bc.cfg = cfg

	// If startingBalances was specified, we initialize our balances to that object.
	balances := startingBalances
	if balances == nil {
		balances = make(map[string]int)
	}

	// If clientBalanceMap was initialized instead, we copy over those values.
	if clientBalanceMap != nil {
		for client, balance := range clientBalanceMap {
			balances[client.address] = balance
		}
	}

	g := makeBlock("", nil, bc.POW_TARGET(), bc.COINBASE_AMT_ALLOWED())

	// Initializing starting balances in the genesis block.
	for addr, balance := range balances {
		g.balances[addr] = balance
	}

	// If clientBalanceMap was specified, we set the genesis block for every client.
	if clientBalanceMap != nil {
		for client := range clientBalanceMap {
			client.setGenesisBlock(g)
		}
	}

	return g
}
func makeBlock(rewardAddr string, prevBlock *Block, target *big.Int, coinbaseReward int) Block {
	return Block{rewardAddr: rewardAddr, prevBlock: prevBlock, target: target, coinbaseReward: coinbaseReward}
}

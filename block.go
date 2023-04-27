package blockchain

// import "math/big"
import (
	"encoding/json"
	"fmt"
	"math/big"
	"spartangold/utils"
	"time"
)

/**
 * A block is a collection of transactions, with a hash connecting it
 * to a previous block.
 */

type Block struct {
	rewardAddr string
	// prevBlock      *Block
	prevBlockHash  string
	target         *big.Int //was Number in JS
	coinbaseReward int      //was number in JS
	balances       map[string]int
	nextNonce      map[string]int //confirm
	transactions   map[string]Transaction
	chainLength    int
	timestamp      int64
}

/**
 * Creates a new Block.  Note that the previous block will not be stored;
 * instead, its hash value will be maintained in this block.
 *
 * @constructor
 * @param {String} rewardAddr - The address to receive all mining rewards for this block.
 * @param {Block} [prevBlock] - The previous block in the blockchain.
 * @param {Number} [target] - The POW target.  The miner must find a proof that
 *      produces a smaller value when hashed.
 * @param {Number} [coinbaseReward] - The gold that a miner earns for finding a block proof.
 */

func newBlock(rewardAddr string, prevBlock *Block, target big.Int, coinbaseReward int) Block {
	var block Block
	// block.prevBlock = prevBlock
	block.target = &target

	if prevBlock != nil {
		block.prevBlockHash = prevBlock.hashVal()
		// Get the balances and nonces from the previous block, if available.
		// Note that balances and nonces are NOT part of the serialized format.
		block.balances = make(map[string]int, len(prevBlock.balances))
		for k, v := range prevBlock.balances {
			block.balances[k] = v
		}
		block.nextNonce = make(map[string]int)
		for k, v := range prevBlock.nextNonce {
			block.nextNonce[k] = v
		}
		if prevBlock.rewardAddr != "nil" {
			var winnerBalance int
			winnerBalance = block.balanceOf(prevBlock.rewardAddr)
			block.balances[prevBlock.rewardAddr] = winnerBalance + prevBlock.totalRewards()

		}
	} else {
		block.prevBlockHash = "nil"
		block.balances = make(map[string]int)
		block.nextNonce = make(map[string]int)
	}
	// Storing transactions in a Map to preserve key order.
	block.transactions = make(map[string]Transaction)

	// Used to determine the winner between competing chains.
	// Note that this is a little simplistic -- an attacker
	// could make a long, but low-work chain.  However, this works
	// well enough for us.
	if prevBlock != nil {
		block.chainLength = prevBlock.chainLength
	} else {
		block.chainLength = 0
	}

	block.timestamp = time.Now().UnixNano() / int64(time.Millisecond)

	// The address that will gain both the coinbase reward and transaction fees,
	// assuming that the block is accepted by the network.
	block.rewardAddr = rewardAddr
	block.coinbaseReward = coinbaseReward

	return block
}

/**
 * Determines whether the block is the beginning of the chain.
 *
 * @returns {bool} - True if this is the first block in the chain.
 */
func (b *Block) isGenesisBlock() bool {
	return b.chainLength == 0
}

/**
 * Returns true if the hash of the block is less than the target
 * proof of work value.
 *
 * @returns {Boolean} - True if the block has a valid proof.
 */
func (b *Block) hasvalidProof() bool {
	h := utils.Hash(b.serialize())
	var n, _ = new(big.Int).SetString("0x"+h, 0)
	cmp := n.Cmp(b.target)
	return cmp < 0
}

/**
 * Converts a Block into string form.  Some fields are deliberately omitted.
 * Note that Block.deserialize plus block.rerun should restore the block.
 *
 * @returns {String} - The block in JSON format.
 */
func (b *Block) serialize() string {
	jsonBytes, err := json.Marshal(b)
	if err != nil {
		fmt.Println(err)
	}
	return string(jsonBytes)
}

func (b *Block) hashVal() string {
	return utils.Hash(b.serialize())
}

/**
 * Returns the hash of the block as its id.
 *
 * @returns {String} - A unique ID for the block.
 */
func (b *Block) id() string {
	return b.hashVal()
}

func (b *Block) balanceOf(rewardAddr string) int {
	//TODO
	if rewardAddr != "nil" {
		return 0 //TODO implement
	} else {
		return 0
	}

}

func (b *Block) totalRewards() int {
	//TODO
	return 0
}

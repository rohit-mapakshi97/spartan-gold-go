package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"spartangold/utils"
	"sync"

	"github.com/chuckpreslar/emission"
)

/**
 * A client has a public/private keypair and an address.
 * It can send and receive messages on the Blockchain network.
 */
type Client struct {
	Name                        string
	Address                     string
	PrivKey                     *rsa.PrivateKey
	PubKey                      *rsa.PublicKey
	Blocks                      map[string]*Block
	PendingOutgoingTransactions map[string]*Transaction
	PendingReceivedTransactions map[string]*Transaction
	PendingBlocks               map[string]*utils.Set[*Block]
	LastBlock                   *Block
	LastConfirmedBlock          *Block
	ReceivedBlock               *Block
	Nonce                       uint32
	Net                         *FakeNet
	Emitter                     *emission.Emitter
	mu                          sync.Mutex
}

type Message struct {
	Address       string
	PrevBlockHash string
}

func NewClient(name string, Net *FakeNet, startingBlock *Block) *Client {
	var c Client
	c.Net = Net
	c.Name = name

	c.PrivKey, c.PubKey = utils.GenerateKeypair()

	c.Address = utils.CalcAddress(c.PubKey)
	// Establishes order of transactions.  Incremented with each
	// new output transaction from this client.  This feature
	// avoids replay attacks.
	c.Nonce = 0

	// A map of transactions where the client has spent money,
	// but where the transaction has not yet been confirmed.
	c.PendingOutgoingTransactions = make(map[string]*Transaction)

	// A map of transactions received but not yet confirmed.
	c.PendingReceivedTransactions = make(map[string]*Transaction)

	// A map of all block hashes to the accepted blocks.
	c.Blocks = make(map[string]*Block)

	// A map of missing block IDS to the list of blocks depending
	// on the missing blocks.
	c.PendingBlocks = make(map[string]*utils.Set[*Block])

	if startingBlock != nil {
		c.SetGenesisBlock(startingBlock)
	}

	// Setting up listeners to receive messages from other clients.
	c.Emitter = emission.NewEmitter()
	c.Emitter.On(PROOF_FOUND, c.ReceiveBlockBytes)
	c.Emitter.On(MISSING_BLOCK, c.ProvideMissingBlock)
	return &c
}

// The genesis block can only be set if the client does not already have the genesis block.
func (c *Client) SetGenesisBlock(startingBlock *Block) {

	if (*c).LastBlock != nil {
		fmt.Printf("Cannot set starting block for existing blockchain.")
	}
	// Transactions from this block or older are assumed to be confirmed,
	// and therefore are spendable by the client. The transactions could
	// roll back, but it is unlikely.
	(*c).LastConfirmedBlock = startingBlock

	// The last block seen.  Any transactions after lastConfirmedBlock
	// up to lastBlock are considered pending.
	(*c).LastBlock = startingBlock
	blockId := startingBlock.GetHash()
	(*c).Blocks[blockId] = startingBlock
}

// The amount of gold available to the client looking at the last confirmed block
func (c *Client) ConfirmedBalance() uint32 {
	return (*c).LastConfirmedBlock.BalanceOf((*c).Address)
}

/**
 * Any gold received in the last confirmed block or before is considered
 * spendable, but any gold received more recently is not yet available.
 * However, any gold given by the client to other clients in unconfirmed
 * transactions is treated as unavailable.
 */
func (c *Client) AvailableGold() uint32 {
	var pendingSpent uint32 = 0
	for _, tx := range (*c).PendingOutgoingTransactions {
		pendingSpent += tx.TotalOutput()
	}
	return c.ConfirmedBalance() - pendingSpent
}

/**
 * Broadcasts a transaction from the client giving gold to the clients
 * specified in 'outputs'. A transaction fee may be specified, which can
 * be more or less than the default value.*/
func (c *Client) PostTransaction(outputs []Output, fee uint32) *Transaction {

	(*c).mu.Lock()
	defer (*c).mu.Unlock()

	total := fee
	for _, output := range outputs {
		total += output.Amount
	}
	// Make sure the client has enough gold.
	if total > c.AvailableGold() {
		panic(`Account doesn't have enough balance for transaction`)
	}
	tx := NewTransaction((*c).Address, (*c).Nonce, (*c).PubKey, nil, fee, outputs, nil)

	tx.Sign((*c).PrivKey)
	(*c).PendingOutgoingTransactions[tx.Id()] = tx
	(*c).Nonce++
	data := TransactionToBytes(tx)
	// Create and broadcast the transaction.
	(*c).Net.Broadcast(POST_TRANSACTION, data)

	return tx
}

/**
 * Validates and adds a block to the list of blocks, possibly updating the head
 * of the blockchain.  Any transactions in the block are rerun in order to
 * update the gold balances for all clients.  If any transactions are found to be
 * invalid due to lack of funds, the block is rejected and 'null' is returned to
 * indicate failure.
 *
 * If any blocks cannot be connected to an existing block but seem otherwise valid,
 * they are added to a list of pending blocks and a request is sent out to get the
 * missing blocks from other clients.*/
func (c *Client) ReceiveBlock(b Block) *Block {
	(*c).mu.Lock()
	defer (*c).mu.Unlock()

	block := &b
	blockId := block.GetHash()

	if _, received := (*c).Blocks[blockId]; received {
		return nil
	}

	if !block.hasValidProof() && !block.IsGenesisBlock() {
		c.Log(fmt.Sprintf("Block %v does not have a valid proof\n", blockId))
		return nil
	}

	//var prevBlock *Block = nil
	prevBlock, received := (*c).Blocks[(*block).PrevBlockHash]
	if !received && !block.IsGenesisBlock() {

		stuckBlocks, received := (*c).PendingBlocks[(*block).PrevBlockHash]
		if !received {
			c.RequestMissingBlock(block)
			// TODO: Define a set
			stuckBlocks = utils.NewSet[*Block]()
		}
		stuckBlocks.Add(block)
		(*c).PendingBlocks[(*block).PrevBlockHash] = stuckBlocks
		return nil

	}

	if !block.IsGenesisBlock() {
		if !block.Rerun(prevBlock) {
			return nil
		}
	}

	blockId = block.GetHash()
	(*c).Blocks[blockId] = block

	if (*(*c).LastBlock).ChainLength < (*block).ChainLength {
		(*c).LastBlock = block
		c.SetLastConfirmed()
	}

	unstuckBlocks, received := (*c).PendingBlocks[blockId]
	var unstuckBlocksArr []*Block
	if received {
		unstuckBlocksArr = unstuckBlocks.ToArray()
	}

	delete((*c).PendingBlocks, blockId)

	for _, uBlock := range unstuckBlocksArr {
		c.Log(fmt.Sprintf("processing unstuck block %v", uBlock.GetHashStr()))
		// Need to change the "" into empty []byte
		go c.ReceiveBlock(*uBlock)
	}
	c.Log(fmt.Sprintf("block %s received", block.GetHashStr()))
	return block
}

func (c *Client) ReceiveBlockBytes(bs []byte) *Block {

	block := BytesToBlock(bs)
	return c.ReceiveBlock(*block)
}

// Request the previous block from the network.
func (c *Client) RequestMissingBlock(block *Block) {
	c.Log(fmt.Sprintf("Asking for missing block: %v", (*block).PrevBlockHash))
	var msg = Message{(*c).Address, (*block).PrevBlockHash}
	jsonByte, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("RequestMissingBlock() Marshal Panic:")
		panic(err)
	}
	(*c).Net.Broadcast(MISSING_BLOCK, jsonByte)
}

/**
 * Resend any transactions in the pending list.
 */
func (c *Client) ResendPendingTransactions() {
	for _, tx := range (*c).PendingOutgoingTransactions {
		jsonByte, err := json.Marshal(*tx)
		if err != nil {
			fmt.Println("ResendPendingTransactions() Marshal Panic:")
			panic(err)
		}
		(*c).Net.Broadcast(POST_TRANSACTION, jsonByte)
	}
}

/**
 * Takes an object representing a request for a missing block.
 * If the client has the block, it will send the block to the
 * client that requested it.*/
func (c *Client) ProvideMissingBlock(data []byte) {
	(*c).mu.Lock()
	defer (*c).mu.Unlock()
	var msg Message
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println("ProvideMissingBlock() unmarshal Panic:")
		panic(err)
	}
	if val, received := (*c).Blocks[msg.PrevBlockHash]; received {
		c.Log(fmt.Sprintf("Providing missing block %v", val.GetHashStr()))
		data := BlockToBytes(val)
		(*c).Net.SendMessage(msg.Address, PROOF_FOUND, data)
	}
}

/**
 * Sets the last confirmed block according to the most recently accepted block,
 * also updating pending transactions according to this block.
 * Note that the genesis block is always considered to be confirmed.
 */
func (c *Client) SetLastConfirmed() {
	block := (*c).LastBlock
	confirmedBlockHeight := uint32(0)
	if (*block).ChainLength > CONFIRMED_DEPTH {
		confirmedBlockHeight = (*block).ChainLength - CONFIRMED_DEPTH
	}
	for (*block).ChainLength > confirmedBlockHeight {
		block = (*c).Blocks[(*block).PrevBlockHash]
	}
	(*c).LastConfirmedBlock = block
	// Update pending transactions according to the new last confirmed block.
	for id, tx := range (*c).PendingOutgoingTransactions {
		if (*c).LastConfirmedBlock.Contains(tx) {
			delete((*c).PendingOutgoingTransactions, id)
		}
	}
}

// Utility method that displays all confirmed balances for all clients
func (c *Client) ShowAllBalances() {

	fmt.Printf("Showing balances:")
	for id, balance := range (*(*c).LastConfirmedBlock).Balances {
		fmt.Printf("	%v", id)
		fmt.Printf("	%v", balance)
		fmt.Println("")
	}
}

// Logs messages to stdout
func (c *Client) Log(msg string) {
	name := (*c).Address[0:10]
	if len((*c).Name) > 0 {
		name = (*c).Name
	}
	fmt.Printf("	%s", name)
	fmt.Printf("	%s\n", msg)
}

// Print out the blocks in the blockchain from the current head to the genesis block.
func (c *Client) ShowBlockchain() {

	block := (*c).LastBlock
	fmt.Println("BLOCKCHAIN:")
	for block != nil {
		blockId := block.GetHash()
		fmt.Println(blockId)
		block = (*c).Blocks[(*block).PrevBlockHash]
	}
}

func (c *Client) GetAddress() string {
	return (*c).Address
}
func (c *Client) GetEmitter() *emission.Emitter {
	return (*c).Emitter
}

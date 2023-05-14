package main

import (
	"fmt"
	"time"
)

func main() {
	net := NewFakeNet()

	// Clients
	alice := NewClient("Alice", net, nil)
	bob := NewClient("Bob", net, nil)
	cindy := NewClient("Cindy", net, nil)
	// Miners
	minnie := NewMiner("Minnie", net, NUM_ROUNDS_MINING, nil)
	mickey := NewMiner("Mickey", net, NUM_ROUNDS_MINING, nil)

	// Creating genesis block
	initialBalances := make(map[string]uint32)
	initialBalances[alice.GetAddress()] = 233
	initialBalances[bob.GetAddress()] = 99
	initialBalances[cindy.GetAddress()] = 67
	initialBalances[minnie.GetAddress()] = 400
	initialBalances[mickey.GetAddress()] = 300

	genesis := MakeGenesisDefault(initialBalances)

	// Late Miner
	donald := NewMiner("Donald", net, NUM_ROUNDS_MINING, genesis)

	// Setting genesis block for other clients and miners
	//TODO implement inheritance between Miner and client and set this during Genesis block creation
	alice.SetGenesisBlock(genesis)
	bob.SetGenesisBlock(genesis)
	cindy.SetGenesisBlock(genesis)
	minnie.SetGenesisBlock(genesis)
	mickey.SetGenesisBlock(genesis)

	printClientBalances := func(c *Client) {
		fmt.Printf("Alice has %d gold\n", c.LastBlock.BalanceOf(alice.GetAddress()))
		fmt.Printf("Bob has %d gold\n", c.LastBlock.BalanceOf(bob.GetAddress()))
		fmt.Printf("Cindy has %d gold\n", c.LastBlock.BalanceOf(cindy.GetAddress()))
		fmt.Printf("Minnie has %d gold\n", c.LastBlock.BalanceOf(minnie.GetAddress()))
		fmt.Printf("Mickey has %d gold\n", c.LastBlock.BalanceOf(mickey.GetAddress()))
		fmt.Printf("Donald has %d gold\n", c.LastBlock.BalanceOf(donald.GetAddress()))
	}

	printMinerBalance := func(m *Miner) {
		fmt.Printf("Alice has %d gold\n", m.LastBlock.BalanceOf(alice.GetAddress()))
		fmt.Printf("Bob has %d gold\n", m.LastBlock.BalanceOf(bob.GetAddress()))
		fmt.Printf("Cindy has %d gold\n", m.LastBlock.BalanceOf(cindy.GetAddress()))
		fmt.Printf("Minnie has %d gold\n", m.LastBlock.BalanceOf(minnie.GetAddress()))
		fmt.Printf("Mickey has %d gold\n", m.LastBlock.BalanceOf(mickey.GetAddress()))
		fmt.Printf("Donald has %d gold\n", m.LastBlock.BalanceOf(donald.GetAddress()))
	}

	// Showing the initial balances from Alice's perspective, for no particular reason.
	fmt.Printf("Initial balances:")
	printClientBalances(alice)

	net.Register(alice, bob, cindy, minnie, mickey)

	// Miners start mining.
	minnie.Initialize()
	mickey.Initialize()

	// Alice transfers some money to Bob.
	output1 := Output{Address: bob.GetAddress(), Amount: 40}
	outputs := []Output{output1}

	alice.PostTransaction(outputs, DEFAULT_TX_FEE)

	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println()
		fmt.Println("***Starting a late-to-the-party miner***")
		fmt.Println()
		net.Register(donald)
		donald.Initialize()
	}()

	// Print out the final balances after it has been running for some time.
	time.Sleep(5 * time.Second)
	fmt.Println()
	fmt.Printf("Minnie has a chain of length %d\n", minnie.CurrentBlock.ChainLength)

	fmt.Println()
	fmt.Printf("Mickey has a chain of length %d\n", mickey.CurrentBlock.ChainLength)

	fmt.Println()
	fmt.Printf("Donald has a chain of length %d\n", donald.CurrentBlock.ChainLength)

	fmt.Println()
	fmt.Println("Final Balances (Minnie's perspective):")
	printMinerBalance(minnie)

	fmt.Println()
	fmt.Println("Final Balances (Alice's perspective):")
	printClientBalances(alice)

	fmt.Println()
	fmt.Println("Final Balances (Donald's perspective):")
	printMinerBalance(donald)

	alice.ShowBlockchain()
	fmt.Println("End!")
}

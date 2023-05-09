package blockchain

// Constants for mining
const NUM_ROUNDS_MINING = 2000

type Miner struct {
	Client
	miningRounds int
}

func NewMiner(options map[string]interface{}) *Miner {
	m := &Miner{}
	k := NewClient(map[string]interface{}{
		"name": options["name"],
	})
	m.name = k.name
	m.address = k.address
	m.publicKey = k.publicKey
	m.privateKey = k.privateKey
	m.miningRounds = NUM_ROUNDS_MINING

	return m
}

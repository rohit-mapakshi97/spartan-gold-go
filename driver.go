package blockchain

import (
	"fmt"
)

func main() {
	fmt.Println("Starting simulation...")
	// Client alice;
	alice := NewClient(map[string]interface{}{
		"name": "Alice",
	})
	bob := NewClient(map[string]interface{}{
		"name": "Bob",
	})
	charlie := NewClient(map[string]interface{}{
		"name": "Charlie",
	})
	fmt.Println(alice.name)
	fmt.Println(bob.name)
	fmt.Println(charlie.name)

	minnie := NewMiner(map[string]interface{}{
		"name": "Minnie",
	})
	mickey := NewMiner(map[string]interface{}{
		"name": "Mickey",
	})

	fmt.Println(minnie.name)
	fmt.Println(mickey.name)

	genesis := makeGenesis(map[string]int{
		"alice":   233,
		"bob":     99,
		"charlie": 67,
		"minnie":  400,
		"mickey":  300,
	})

}

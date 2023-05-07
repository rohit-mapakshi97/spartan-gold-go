package blockchain

type FakeNet struct {
	clients map[string]interface{}
}

func (fn *FakeNet) Register(clientList ...*Client) {
	for _, client := range clientList {
		fn.clients[client.address] = client
	}
}

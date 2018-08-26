package chord

type ContactInfo struct {
	Address string `json:"address"`
	Id      NodeID `json:"id"`
	Payload []byte `json:"payload"`
}
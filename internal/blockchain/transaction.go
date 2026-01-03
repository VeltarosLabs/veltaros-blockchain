package blockchain

// Transaction represents a value transfer
type Transaction struct {
	ID        string
	From      string
	To        string
	Amount    int
	Timestamp int64
}

package blockchain

type Block struct {
	Index        int
	Timestamp    int64
	Transactions []Transaction
	PrevHash     string
	Hash         string
	Nonce        int

	UTXOTxs []UTXOTransaction `json:"utxo_txs,omitempty"`
}

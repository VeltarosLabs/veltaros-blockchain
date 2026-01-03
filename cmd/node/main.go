package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
	"github.com/VeltarosLabs/veltaros-blockchain/internal/p2p"
	"github.com/VeltarosLabs/veltaros-blockchain/internal/wallet"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "start":
		startNode()
	case "wallet-new":
		createWallet()
	case "tx":
		sendSignedTransaction()
	case "mine":
		mineLocalDemo()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Print(
		`Veltaros Blockchain Node

Commands:
  start --port 3000 [--peer localhost:3001]     Start P2P node
  wallet-new --out wallet.pem                   Create wallet
  tx --wallet wallet.pem --to Bob --amount 10 --node localhost:3000
                                                Send signed transaction
  mine                                          Local demo mining (temporary)
`)
}

func startNode() {
	cmd := flag.NewFlagSet("start", flag.ExitOnError)
	port := cmd.String("port", "3000", "Node port")
	peer := cmd.String("peer", "", "Peer address (optional)")
	cmd.Parse(os.Args[2:])

	bc := blockchain.NewBlockchain()
	node := p2p.NewNode(":"+*port, bc)

	go func() {
		if err := node.StartServer(); err != nil {
			fmt.Println("Server error:", err)
		}
	}()

	time.Sleep(300 * time.Millisecond)

	if *peer != "" {
		if err := node.Connect(*peer); err != nil {
			fmt.Println("Failed to connect:", err)
		}
	}

	fmt.Println("Node running on port", *port, "(CTRL+C to stop)")
	select {}
}

func createWallet() {
	cmd := flag.NewFlagSet("wallet-new", flag.ExitOnError)
	out := cmd.String("out", "wallet.pem", "Output wallet PEM file")
	cmd.Parse(os.Args[2:])

	w, err := wallet.New()
	if err != nil {
		fmt.Println("Wallet error:", err)
		return
	}

	if err := wallet.SavePrivateKeyPEM(*out, w.PrivateKey); err != nil {
		fmt.Println("Save error:", err)
		return
	}

	fmt.Println("Wallet created!")
	fmt.Println("Address:", w.Address)
	fmt.Println("Saved private key to:", *out)
}

func sendSignedTransaction() {
	cmd := flag.NewFlagSet("tx", flag.ExitOnError)
	walletPath := cmd.String("wallet", "wallet.pem", "Wallet PEM file")
	to := cmd.String("to", "", "Recipient address")
	amount := cmd.Int("amount", 0, "Amount")
	nodeAddr := cmd.String("node", "localhost:3000", "Node address host:port")
	cmd.Parse(os.Args[2:])

	if *to == "" || *amount <= 0 {
		fmt.Println("Invalid tx args. Example: tx --wallet wallet.pem --to <addr> --amount 10 --node localhost:3000")
		return
	}

	priv, err := wallet.LoadPrivateKeyPEM(*walletPath)
	if err != nil {
		fmt.Println("Load wallet error:", err)
		return
	}

	// Rebuild wallet object
	pubBytes := ellipticMarshalPublic(&priv.PublicKey)
	fromAddr := wallet.AddressFromPublicKey(pubBytes)

	w := &wallet.Wallet{
		PrivateKey: priv,
		PublicKey:  pubBytes,
		Address:    fromAddr,
	}

	tx := blockchain.NewUnsignedTransaction(w.Address, *to, *amount)
	if err := tx.SignWith(w); err != nil {
		fmt.Println("Sign error:", err)
		return
	}

	if err := sendTxToNode(*nodeAddr, tx); err != nil {
		fmt.Println("Send error:", err)
		return
	}

	fmt.Println("Signed transaction sent to node:", *nodeAddr)
	fmt.Println("From:", tx.From)
	fmt.Println("To:", tx.To)
	fmt.Println("Amount:", tx.Amount)
}

func sendTxToNode(addr string, tx blockchain.Transaction) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	raw, _ := json.Marshal(tx)
	msg := p2p.Message{
		Type: p2p.MsgTransaction,
		Data: raw,
	}

	return json.NewEncoder(conn).Encode(msg)
}

// Temporary demo mining (until we add remote mine RPC)
func mineLocalDemo() {
	bc := blockchain.NewBlockchain()
	bc.MinePendingTransactions()
	fmt.Println("Mined locally (demo). Chain length:", len(bc.Blocks))
}

// Local helper to avoid importing elliptic in many places
func ellipticMarshalPublic(pub interface{}) []byte {
	// pub is *ecdsa.PublicKey, but kept generic to keep this file small.
	// We'll do the proper conversion here with minimal imports:
	// We intentionally inline a tiny marshal to avoid exposing wallet internals in CLI.

	// This is safe because wallet.go uses P-256 and Marshal format is standard.
	pk, ok := pub.(*ecdsa.PublicKey)
	if !ok || pk == nil {
		return nil
	}
	return elliptic.Marshal(elliptic.P256(), pk.X, pk.Y)
}

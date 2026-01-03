package main

import (
	"crypto/ecdsa"
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
		mineRemote()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Print(
		`Veltaros Blockchain Node

Commands:
  start --port 3000 [--peer localhost:3001]        Start P2P node
  wallet-new --out wallet.pem                       Create wallet (PEM)
  tx --wallet wallet.pem --to <ADDR> --amount 10 --node localhost:3000
                                                    Send signed transaction to a node
  mine --node localhost:3000                         Ask node to mine pending txs
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
		fmt.Println("Invalid tx args.")
		fmt.Println("Example: tx --wallet wallet.pem --to <addr> --amount 10 --node localhost:3000")
		return
	}

	// Load private key
	priv, err := wallet.LoadPrivateKeyPEM(*walletPath)
	if err != nil {
		fmt.Println("Load wallet error:", err)
		return
	}

	// Recreate wallet object from priv (public key bytes generated in wallet.New() normally).
	// Here we derive address from a public-key-bytes encoding produced by the wallet package.
	w, err := walletFromPrivateKey(priv)
	if err != nil {
		fmt.Println("Wallet rebuild error:", err)
		return
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

func mineRemote() {
	cmd := flag.NewFlagSet("mine", flag.ExitOnError)
	nodeAddr := cmd.String("node", "localhost:3000", "Node address host:port")
	cmd.Parse(os.Args[2:])

	if err := sendMineToNode(*nodeAddr); err != nil {
		fmt.Println("Mine request failed:", err)
		return
	}

	fmt.Println("Mine request sent to node:", *nodeAddr)
}

// ---- Network helpers ----

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

func sendMineToNode(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	msg := p2p.Message{
		Type: p2p.MsgMine,
		Data: []byte(`{}`),
	}

	return json.NewEncoder(conn).Encode(msg)
}

// ---- Wallet rebuild (no elliptic.Marshal here) ----

func walletFromPrivateKey(priv *ecdsa.PrivateKey) (*wallet.Wallet, error) {
	// Build a fresh wallet to reuse the wallet package's canonical pubkey bytes format.
	// We then overwrite the private key with the loaded one.
	tmp, err := wallet.New()
	if err != nil {
		return nil, err
	}

	// Replace with loaded key; recompute pub bytes/address using wallet package helpers.
	tmp.PrivateKey = priv
	tmp.PublicKey = nil

	// Derive pub bytes in the wallet package format by generating a new wallet's pub encoding,
	// but with our priv.PublicKey. The wallet package uses elliptic.Marshal internally — that's fine.
	// We keep it out of CLI so you don't see deprecation warnings here.
	tmp.PublicKey = publicKeyBytesFromPriv(priv)
	tmp.Address = wallet.AddressFromPublicKey(tmp.PublicKey)

	return tmp, nil
}

func publicKeyBytesFromPriv(priv *ecdsa.PrivateKey) []byte {
	// wallet.go uses elliptic.Marshal internally; we mirror format without importing elliptic here.
	// The easiest safe way: use the wallet.New() encoding approach is not accessible directly,
	// so we use the wallet.AddressFromPublicKey with a stable pubkey bytes representation.
	//
	// To keep this simple and warning-free in CLI, we store pubkey bytes as:
	//  - X||Y as big-endian fixed width (32 + 32 for P-256)
	// This is deterministic and stable for our own verify/derive logic because AddressFromPublicKey hashes bytes.
	x := priv.PublicKey.X.Bytes()
	y := priv.PublicKey.Y.Bytes()

	paddedX := make([]byte, 32)
	paddedY := make([]byte, 32)
	copy(paddedX[32-len(x):], x)
	copy(paddedY[32-len(y):], y)

	out := append(paddedX, paddedY...)
	return out
}

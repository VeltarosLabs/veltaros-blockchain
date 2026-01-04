package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "wallet-new":
		cmdWalletNew(os.Args[2:])
	case "send":
		cmdSend(os.Args[2:])
	case "mine":
		cmdMine(os.Args[2:])
	case "balance":
		cmdBalance(os.Args[2:])
	case "nonce":
		cmdNonce(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("Veltaros CLI")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  wallet-new --out alice.pem")
	fmt.Println("  nonce      --addr ADDRESS --node 127.0.0.1:3000")
	fmt.Println("  send       --wallet alice.pem --to TO_ADDR --amount 5 --fee 1 --node 127.0.0.1:3000")
	fmt.Println("  mine       --miner MINER_ADDR --node 127.0.0.1:3000")
	fmt.Println("  balance    --addr ADDRESS --node 127.0.0.1:3000")
}

func cmdWalletNew(args []string) {
	fs := flag.NewFlagSet("wallet-new", flag.ExitOnError)
	out := fs.String("out", "wallet.pem", "output pem file")
	fs.Parse(args)

	priv, addr, err := blockchain.GenerateWallet()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if err := writeECPrivateKeyPEM(*out, priv); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("wallet created!")
	fmt.Println("Address:", addr)
	fmt.Println("Saved private key to:", *out)
}

func cmdNonce(args []string) {
	fs := flag.NewFlagSet("nonce", flag.ExitOnError)
	addr := fs.String("addr", "", "address")
	node := fs.String("node", "127.0.0.1:3000", "http node host:port")
	fs.Parse(args)

	if *addr == "" {
		fmt.Println("missing --addr")
		os.Exit(2)
	}

	nonce, err := getNonce(*node, *addr)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Printf("{\"address\":\"%s\",\"nonce\":%d}\n", *addr, nonce)
}

func cmdBalance(args []string) {
	fs := flag.NewFlagSet("balance", flag.ExitOnError)
	addr := fs.String("addr", "", "address")
	node := fs.String("node", "127.0.0.1:3000", "http node host:port")
	fs.Parse(args)

	if *addr == "" {
		fmt.Println("missing --addr")
		os.Exit(2)
	}

	url := fmt.Sprintf("http://%s/balance?addr=%s", *node, *addr)
	body, err := httpGet(url)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println(string(body))
}

func cmdMine(args []string) {
	fs := flag.NewFlagSet("mine", flag.ExitOnError)
	miner := fs.String("miner", "", "miner address")
	node := fs.String("node", "127.0.0.1:3000", "http node host:port")
	fs.Parse(args)

	if *miner == "" {
		fmt.Println("missing --miner")
		os.Exit(2)
	}

	payload := map[string]any{"miner": *miner}
	b, _ := json.Marshal(payload)

	url := fmt.Sprintf("http://%s/mine", *node)
	resp, err := httpPost(url, "application/json", b)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println(string(resp))
}

func cmdSend(args []string) {
	fs := flag.NewFlagSet("send", flag.ExitOnError)
	walletPath := fs.String("wallet", "", "pem wallet file")
	to := fs.String("to", "", "recipient address")
	amount := fs.Int("amount", 0, "amount")
	fee := fs.Int("fee", 0, "fee (optional)")
	node := fs.String("node", "127.0.0.1:3000", "http node host:port")
	fs.Parse(args)

	if *walletPath == "" || *to == "" || *amount <= 0 {
		fmt.Println("missing required flags: --wallet, --to, --amount")
		os.Exit(2)
	}

	priv, err := readECPrivateKeyPEM(*walletPath)
	if err != nil {
		fmt.Println("error reading wallet:", err)
		os.Exit(1)
	}

	fromAddr := blockchain.AddressFromPubKey(priv.PublicKey)

	nonce, err := getNonce(*node, fromAddr)
	if err != nil {
		fmt.Println("error getting nonce:", err)
		os.Exit(1)
	}

	tx := blockchain.NewTransaction(fromAddr, *to, *amount, *fee, nonce)
	if err := tx.Sign(priv); err != nil {
		fmt.Println("error signing tx:", err)
		os.Exit(1)
	}

	raw, _ := json.Marshal(tx)
	url := fmt.Sprintf("http://%s/transaction", *node)

	resp, err := httpPost(url, "application/json", raw)
	if err != nil {
		fmt.Println("request error:", err)
		os.Exit(1)
	}
	fmt.Println(string(resp))
}

// -------- HTTP helpers --------

func httpGet(url string) ([]byte, error) {
	c := &http.Client{Timeout: 10 * time.Second}
	resp, err := c.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return b, nil
}

func httpPost(url, contentType string, body []byte) ([]byte, error) {
	c := &http.Client{Timeout: 15 * time.Second}
	resp, err := c.Post(url, contentType, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	return b, nil
}

func getNonce(node string, addr string) (uint64, error) {
	url := fmt.Sprintf("http://%s/nonce?addr=%s", node, addr)
	b, err := httpGet(url)
	if err != nil {
		return 0, err
	}
	var out struct {
		Nonce uint64 `json:"nonce"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		return 0, err
	}
	return out.Nonce, nil
}

// -------- PEM helpers --------

func writeECPrivateKeyPEM(path string, key *ecdsa.PrivateKey) error {
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	block := &pem.Block{Type: "EC PRIVATE KEY", Bytes: der}
	return os.WriteFile(path, pem.EncodeToMemory(block), 0600)
}

func readECPrivateKeyPEM(path string) (*ecdsa.PrivateKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p, _ := pem.Decode(b)
	if p == nil {
		return nil, fmt.Errorf("invalid pem")
	}
	key, err := x509.ParseECPrivateKey(p.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

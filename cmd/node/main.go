package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/wallet"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "wallet-new":
		walletNew()
	case "send":
		sendTx()
	case "mine":
		mine()
	case "balance":
		balance()
	case "chain":
		chain()
	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Veltaros Node CLI")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  wallet-new  --out <file.pem>                  Create a new wallet private key")
	fmt.Println("  send        --from <addr> --to <addr> --amount <n> --node <host:port>")
	fmt.Println("  mine        --miner <addr> --node <host:port>")
	fmt.Println("  balance     --addr <addr> --node <host:port>")
	fmt.Println("  chain       --node <host:port>")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  go run cmd/node/main.go wallet-new --out alice.pem")
	fmt.Println("  go run cmd/node/main.go wallet-new --out bob.pem")
	fmt.Println("  go run cmd/node/main.go mine --miner ALICE_ADDR --node localhost:3000")
	fmt.Println("  go run cmd/node/main.go send --from ALICE_ADDR --to BOB_ADDR --amount 5 --node localhost:3000")
	fmt.Println("  go run cmd/node/main.go mine --miner ALICE_ADDR --node localhost:3000")
	fmt.Println("  go run cmd/node/main.go balance --addr ALICE_ADDR --node localhost:3000")
	fmt.Println("  go run cmd/node/main.go balance --addr BOB_ADDR --node localhost:3000")
	fmt.Println("  go run cmd/node/main.go chain --node localhost:3000")
}

func walletNew() {
	cmd := flag.NewFlagSet("wallet-new", flag.ExitOnError)
	out := cmd.String("out", "wallet.pem", "output pem file")
	_ = cmd.Parse(os.Args[2:])

	w, err := wallet.NewWallet()
	if err != nil {
		fmt.Println("wallet error:", err)
		return
	}

	if err := wallet.SavePrivateKeyPEM(*out, w.PrivateKey); err != nil {
		fmt.Println("save error:", err)
		return
	}

	fmt.Println("wallet created!")
	fmt.Println("Address:", w.Address)
	fmt.Println("Saved private key to:", *out)
}

// sendTx posts to the node endpoint that accepts tx JSON.
// IMPORTANT: we pass --from address directly (no wallet loading function needed).
func sendTx() {
	cmd := flag.NewFlagSet("send", flag.ExitOnError)
	from := cmd.String("from", "", "sender address")
	to := cmd.String("to", "", "recipient address")
	amount := cmd.Int("amount", 0, "amount to send")
	node := cmd.String("node", "localhost:3000", "node address")
	_ = cmd.Parse(os.Args[2:])

	if *from == "" || *to == "" || *amount <= 0 {
		fmt.Println("Usage: send --from <addr> --to <addr> --amount <n> --node localhost:3000")
		return
	}

	body := map[string]any{
		"from":   *from,
		"to":     *to,
		"amount": *amount,
	}

	b, _ := json.Marshal(body)

	// Your node currently uses /transaction (based on your node.go screenshot)
	resp, err := http.Post("http://"+*node+"/transaction", "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Println("request error:", err)
		return
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Println("node error:", string(data))
		return
	}

	fmt.Println(string(data))
}

func mine() {
	cmd := flag.NewFlagSet("mine", flag.ExitOnError)
	miner := cmd.String("miner", "", "miner address")
	node := cmd.String("node", "localhost:3000", "node address")
	_ = cmd.Parse(os.Args[2:])

	if *miner == "" {
		fmt.Println("Usage: mine --miner <addr> --node localhost:3000")
		return
	}

	body := map[string]any{"miner": *miner}
	b, _ := json.Marshal(body)

	resp, err := http.Post("http://"+*node+"/mine", "application/json", bytes.NewReader(b))
	if err != nil {
		fmt.Println("request error:", err)
		return
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Println("node error:", string(data))
		return
	}
	fmt.Println(string(data))
}

func balance() {
	cmd := flag.NewFlagSet("balance", flag.ExitOnError)
	addr := cmd.String("addr", "", "address")
	node := cmd.String("node", "localhost:3000", "node address")
	_ = cmd.Parse(os.Args[2:])

	if *addr == "" {
		fmt.Println("Usage: balance --addr <addr> --node localhost:3000")
		return
	}

	resp, err := http.Get("http://" + *node + "/balance?addr=" + *addr)
	if err != nil {
		fmt.Println("request error:", err)
		return
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Println("node error:", string(data))
		return
	}

	fmt.Println(string(data))
}

func chain() {
	cmd := flag.NewFlagSet("chain", flag.ExitOnError)
	node := cmd.String("node", "localhost:3000", "node address")
	_ = cmd.Parse(os.Args[2:])

	resp, err := http.Get("http://" + *node + "/chain")
	if err != nil {
		fmt.Println("request error:", err)
		return
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Println("node error:", string(data))
		return
	}

	fmt.Println(string(data))
}

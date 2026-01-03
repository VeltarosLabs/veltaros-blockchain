package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/VeltarosLabs/veltaros-blockchain/internal/blockchain"
	"github.com/VeltarosLabs/veltaros-blockchain/internal/p2p"
)

var (
	port = flag.String("port", "3000", "Node port")
	peer = flag.String("peer", "", "Peer address (optional)")
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "start":
		startNode()

	case "tx":
		sendTransaction()

	case "mine":
		mineBlock()

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Print(
		`Veltaros Blockchain Node

Commands:
  start --port <port> [--peer <addr>]   Start node
  tx --from A --to B --amount N          Send transaction
  mine                                  Mine pending transactions
`)
}

func startNode() {
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)
	startCmd.StringVar(port, "port", "3000", "Node port")
	startCmd.StringVar(peer, "peer", "", "Peer address")

	startCmd.Parse(os.Args[2:])

	bc := blockchain.NewBlockchain()
	node := p2p.NewNode(":"+*port, bc)

	go node.StartServer()
	time.Sleep(time.Second)

	if *peer != "" {
		err := node.Connect(*peer)
		if err != nil {
			fmt.Println("Failed to connect:", err)
		}
	}

	fmt.Println("Node running. Press CTRL+C to exit.")
	select {}
}

func sendTransaction() {
	txCmd := flag.NewFlagSet("tx", flag.ExitOnError)

	from := txCmd.String("from", "", "Sender")
	to := txCmd.String("to", "", "Recipient")
	amount := txCmd.Int("amount", 0, "Amount")

	txCmd.Parse(os.Args[2:])

	if *from == "" || *to == "" || *amount <= 0 {
		fmt.Println("Invalid transaction parameters")
		return
	}

	tx := blockchain.Transaction{
		From:   *from,
		To:     *to,
		Amount: *amount,
	}

	fmt.Println("Transaction created:", tx)
	fmt.Println("NOTE: broadcasting will be added after wallet integration")
}

func mineBlock() {
	bc := blockchain.NewBlockchain()
	bc.MinePendingTransactions()

	fmt.Println("Block mined successfully")
	fmt.Printf("Chain length: %d\n", len(bc.Blocks))
}

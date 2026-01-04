<p align="center">
  <img src="assets/logo.png" alt="Veltaros Logo" width="180" />
</p>

<h1 align="center">Veltaros Blockchain (VLT)</h1>

<p align="center">
  A learning-first blockchain implementation in Go with wallet keys, transactions, mining, HTTP API, and a P2P layer.
</p>

---

## Overview

**Veltaros (VLT)** is a Go-based blockchain project designed to evolve into a fully functional network with:
- Signed transactions (wallet PEM keys)
- Mempool + mining
- Block propagation over P2P
- HTTP API for clients
- Chain sync + longest-valid-chain replacement

> Coin symbol: **VLT**

---

## Repository Layout

- `cmd/veltarosd/` — Veltaros node daemon (HTTP API + P2P)
- `cmd/node/` — CLI tools (wallet, send, mine, balance)
- `internal/blockchain/` — chain, blocks, tx, mining, state
- `internal/wallet/` — key generation + signing
- `internal/network/` — HTTP API server
- `internal/p2p/` — P2P node, peers, messages, server

---

## Quick Start

### 1) Start the node (HTTP + P2P)
Example: HTTP on `3000`, P2P on `4000`

```bash
go run cmd/veltarosd/main.go --addr 3000 --p2p :4000

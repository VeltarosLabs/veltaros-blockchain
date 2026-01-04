# Veltaros Blockchain (VLT)

<img src="assets/logo/veltaros-logo.png" alt="Veltaros (VLT) Logo" width="220" />

**Veltaros (VLT)** is a learning-focused blockchain implementation in Go with:
- Wallet generation (ECDSA)
- Signed transactions
- Mempool + mining
- Simple HTTP API (node)
- P2P sync (blocks + chain sync)

> Unit symbol: **VLT**

---

## Repo Structure

- `cmd/veltarosd/` → Runs the full node (HTTP API + optional P2P)
- `cmd/node/` → CLI client (wallet-new, send, mine, balance)
- `internal/blockchain/` → Core blockchain logic
- `internal/network/` → HTTP API server
- `internal/p2p/` → P2P networking + chain sync
- `assets/` → Branding assets (logo, docs images)

---

## Quick Start (Windows PowerShell)

### 1) Run the full node
Open Terminal 1:

```powershell
go run cmd/veltarosd/main.go --addr 3000 --p2p :4000 --data data

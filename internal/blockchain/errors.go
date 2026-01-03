package blockchain

import "errors"

var (
	ErrInvalidTransaction = errors.New("invalid transaction")
	ErrInvalidBlock       = errors.New("invalid block")
	ErrInsufficientFunds  = errors.New("insufficient funds")
)

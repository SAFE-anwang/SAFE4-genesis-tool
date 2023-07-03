package types

import (
	"math/big"
)

type Safe3Info struct {
	Addr         string     `json:"addr"            gencodec:"required"`
	Amount       *big.Int   `json:"amount"          gencodec:"required"`
	RedeemHeight *big.Int   `json:"redeemHeight"    gencodec:"required"`
}

type Safe3LockInfo struct {
	Addr         string     `json:"addr"            gencodec:"required"`
	Amount       *big.Int   `json:"amount"          gencodec:"required"`
	LockHeight   *big.Int   `json:"lockHeight"      gencodec:"required"`
	UnlockHeight *big.Int   `json:"unlockHeight"    gencodec:"required"`
	Txid         string     `json:"txid"            gencodec:"required"`
	IsMN         bool       `json:"isMN"            gencodec:"required"`
	RedeemHeight *big.Int   `json:"redeemHeight"    gencodec:"required"`
}
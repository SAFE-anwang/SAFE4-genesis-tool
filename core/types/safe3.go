package types

import (
	"github.com/safe/SAFE4-genesis-tool/common"
	"math/big"
)

type Safe3Info struct {
	Safe3Addr           string              `json:"safe3Addr"       gencodec:"required"`
	Amount              *big.Int            `json:"amount"          gencodec:"required"`
	Safe4Addr           common.Address      `json:"safe4Addr"       gencodec:"required"`
	RedeemHeight        *big.Int            `json:"redeemHeight"    gencodec:"required"`
}

type Safe3LockInfo struct {
	Safe3Addr           string              `json:"safe3Addr"       gencodec:"required"`
	Amount              *big.Int            `json:"amount"          gencodec:"required"`
	LockHeight          *big.Int            `json:"lockHeight"      gencodec:"required"`
	UnlockHeight        *big.Int            `json:"unlockHeight"    gencodec:"required"`
	Txid                string              `json:"txid"            gencodec:"required"`
	LockDay             *big.Int            `json:"lockDay"         gencodec:"required"`
	RemainLockHeight    *big.Int            `json:"remainLockHeight" gencodec:"required"`
	IsMN                bool                `json:"isMN"            gencodec:"required"`
	MnState             *big.Int            `json:"MnState"         gencodec:"required"`
	Safe4Addr           common.Address      `json:"safe4Addr"       gencodec:"required"`
	RedeemHeight        *big.Int            `json:"redeemHeight"    gencodec:"required"`
}

type SpecialSafe3Info struct {
	Safe3Addr           string              `json:"safe3Addr"       gencodec:"required"`
	Amount              *big.Int            `json:"amount"          gencodec:"required"`
	Safe4Addr           common.Address      `json:"safe4Addr"       gencodec:"required"`
	ApplyHeight         *big.Int            `json:"applyHeight"     gencodec:"required"`
	Voters              []common.Address    `json:"voters"          gencodec:"required"`
	VoteResults         []*big.Int          `json:"voteResults"     gencodec:"required"`
	RedeemHeight        *big.Int            `json:"redeemHeight"    gencodec:"required"`
}
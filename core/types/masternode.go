package types

import (
	"github.com/safe/SAFE4-genesis-tool/common"
	"math/big"
)

type MasterNodeMemberInfo struct {
	LockID *big.Int         `json:"lockID"    gencodec:"required"`
	Addr   common.Address   `json:"addr"      gencodec:"required"`
	Amount *big.Int         `json:"amount"    gencodec:"required"`
	Height *big.Int         `json:"height"    gencodec:"required"`
}

type MasterNodeIncentivePlan struct {
	Creator *big.Int        `json:"creator"   gencodec:"required"`
	Partner *big.Int        `json:"partner"   gencodec:"required"`
	Voter   *big.Int        `json:"voter"     gencodec:"required"`
}

type MasterNodeInfo struct {
	Id            *big.Int                  `json:"id"            gencodec:"required"`
	Addr          common.Address            `json:"addr"          gencodec:"required"`
	Creator       common.Address            `json:"creator"       gencodec:"required"`
	Amount        *big.Int                  `json:"amount"        gencodec:"required"`
	Enode         string                    `json:"enode"         gencodec:"required"`
	Ip            string                    `json:"ip"            gencodec:"required"`
	Description   string                    `json:"description,omitempty"`
	State               *big.Int            `json:"state"         gencodec:"required"`
	Founders      []MasterNodeMemberInfo    `json:"founders"      gencodec:"required"`
	IncentivePlan MasterNodeIncentivePlan   `json:"incentivePlan" gencodec:"required"`
	CreateHeight  *big.Int                  `json:"createHeight,omitempty"`
	UpdateHeight  *big.Int                  `json:"updateHeight,omitempty"`
}
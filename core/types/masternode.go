package types

import (
	"github.com/safe/SAFE4-genesis-tool/common"
	"math/big"
)

type MasterNodeMemberInfo struct {
	LockID *big.Int				`json:"lockID"`
	Addr   common.Address		`json:"addr"`
	Amount *big.Int             `json:"amount"`
	Height *big.Int				`json:"height"`
}

type MasterNodeIncentivePlan struct {
	Creator *big.Int	`json:"creator"`
	Partner *big.Int	`json:"partner"`
	Voter   *big.Int	`json:"voter"          gencodec:"required"`
}

type MasterNodeInfo struct {
	Id            *big.Int					`json:"id"        gencodec:"required"`
	Addr          common.Address			`json:"addr"        gencodec:"required"`
	Creator       common.Address			`json:"creator"        gencodec:"required"`
	Amount        *big.Int					`json:"amount"        gencodec:"required"`
	Enode         string					`json:"enode"        gencodec:"required"`
	Ip            string					`json:"ip"        gencodec:"required"`
	Description   string					`json:"description"        gencodec:"required"`
	State         *big.Int					`json:"state"        gencodec:"required"`
	Founders      []MasterNodeMemberInfo	`json:"founders"        gencodec:"required"`
	IncentivePlan MasterNodeIncentivePlan   `json:"incentivePlan"        gencodec:"required"`
	CreateHeight  *big.Int					`json:"createHeight"        gencodec:"required"`
	UpdateHeight  *big.Int					`json:"updateHeight"        gencodec:"required"`
}
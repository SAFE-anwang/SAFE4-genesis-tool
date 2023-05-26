package types

import (
	"github.com/safe/SAFE4-genesis-tool/common"
	"math/big"
)

type PropertyInfo struct {
	Name         string
	Value        *big.Int
	Description  string
	CreateHeight *big.Int
	UpdateHeight *big.Int
}

type UnconfirmedPropertyInfo struct {
	Name        string
	Value       *big.Int
	Applicant   common.Address
	Voters      []common.Address
	VoteResults []*big.Int
	Reason      string
	ApplyHeight *big.Int
}
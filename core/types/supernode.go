package types

import (
	"github.com/safe/SAFE4-genesis-tool/common"
	"math/big"
)

type SuperNodeMemberInfo struct {
	LockID *big.Int
	Addr   common.Address
	Amount *big.Int
	Height *big.Int
}

type SuperNodeIncentivePlan struct {
	Creator *big.Int
	Partner *big.Int
	Voter   *big.Int
}

type SuperNodeInfo struct {
	Id               *big.Int
	Name             string
	Addr             common.Address
	Creator          common.Address
	Amount           *big.Int
	Enode            string
	Ip               string
	Description      string
	State            *big.Int
	Founders         []SuperNodeMemberInfo
	IncentivePlan    SuperNodeIncentivePlan
	Voters           []SuperNodeMemberInfo
	TotalVoteNum     *big.Int
	TotalVoterAmount *big.Int
	CreateHeight     *big.Int
	UpdateHeight     *big.Int
}
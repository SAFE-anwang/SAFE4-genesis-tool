package types

import (
	"github.com/safe/SAFE4-genesis-tool/common"
	"math/big"
)

type StateEntry struct {
	Addr   common.Address   `json:"addr"      gencodec:"required"`
	State  *big.Int         `json:"state"     gencodec:"required"`
}
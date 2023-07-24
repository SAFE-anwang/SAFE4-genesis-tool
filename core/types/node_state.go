package types

import (
	"github.com/safe/SAFE4-genesis-tool/common"
	"math/big"
)

type StateEntry struct {
	Addr   common.Address   `json:"addr"      gencodec:"required"`
	State  uint8            `json:"state"     gencodec:"required"`
}

type StateInfo struct {
	Addr   common.Address   `json:"addr"      gencodec:"required"`
	Id     *big.Int         `json:"id"        gencodec:"required"`
	State  uint8            `json:"state"     gencodec:"required"`
}
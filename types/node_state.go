package types

import (
    "github.com/safe/SAFE4-genesis-tool/common"
    "math/big"
)

type StateEntry struct {
    Caller common.Address `json:"caller"    gencodec:"required"`
    State  *big.Int       `json:"state"     gencodec:"required"`
}

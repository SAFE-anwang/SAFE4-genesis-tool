package types

import (
    "github.com/safe/SAFE4-genesis-tool/common"
    "math/big"
)

type LockedData struct {
    Txid             common.Hash
    N                *big.Int
    Amount           *big.Int
    LockHeight       *big.Int
    UnlockHeight     *big.Int
    RemainLockHeight *big.Int
    LockDay          *big.Int
    IsMN             bool
}

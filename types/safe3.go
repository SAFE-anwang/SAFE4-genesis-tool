package types

import (
    "math/big"
)

type LockedData struct {
    Amount           *big.Int
    RemainLockHeight *big.Int
    LockDay          *big.Int
    IsMN             bool
}

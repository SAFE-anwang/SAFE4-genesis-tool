// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
    "github.com/safe/SAFE4-genesis-tool/common"
    "github.com/safe/SAFE4-genesis-tool/common/math"
    "github.com/safe/SAFE4-genesis-tool/params"
)

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.

type Genesis struct {
    Config     *params.ChainConfig   `json:"config"`
    Nonce      string                `json:"nonce"`
    Timestamp  math.HexOrDecimal64   `json:"timestamp"`
    ExtraData  string                `json:"extraData"`
    GasLimit   math.HexOrDecimal64   `json:"gasLimit"   gencodec:"required"`
    Difficulty string                `json:"difficulty" gencodec:"required"`
    Mixhash    common.Hash           `json:"mixHash"`
    Coinbase   common.Address        `json:"coinbase"`
    Alloc      GenesisAlloc          `json:"alloc"      gencodec:"required"`
    Number     string                `json:"number"`
    GasUsed    string                `json:"gasUsed"`
    ParentHash common.Hash           `json:"parentHash"`
    BaseFee    *math.HexOrDecimal256 `json:"baseFeePerGas,omitempty"`
}

type GenesisAlloc map[common.Address]GenesisAccount

// GenesisAccount is an account in the state of the genesis block.
type GenesisAccount struct {
    Balance string                      `json:"balance" gencodec:"required"`
    Code    string                      `json:"code,omitempty"`
    Storage map[common.Hash]common.Hash `json:"storage,omitempty"`
}

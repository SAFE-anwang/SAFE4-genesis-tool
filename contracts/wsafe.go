package contracts

import (
    "github.com/safe/SAFE4-genesis-tool/common"
    "github.com/safe/SAFE4-genesis-tool/types"
    "github.com/safe/SAFE4-genesis-tool/utils"
    "math/big"
    "os"
    "path/filepath"
)

type WSafeStorage struct {
    solcPath     string
    contractPath string
}

func NewWSafeStorage(tool *types.Tool) *WSafeStorage {
    return &WSafeStorage{
        solcPath:     tool.GetSolcPath(),
        contractPath: tool.GetContractPath(),
    }
}

func (s *WSafeStorage) Generate(alloc *types.GenesisAlloc) {
    utils.Compile(s.solcPath, s.contractPath, "WSafe.sol")

    codePath := filepath.Join(s.contractPath, "temp", "WSafe.bin-runtime")
    code, err := os.ReadFile(codePath)
    if err != nil {
        panic(err)
    }

    account := types.GenesisAccount{
        Balance: big.NewInt(0).String(),
        Code:    "0x" + string(code),
    }
    (*alloc)[common.HexToAddress("0x0000000000000000000000000000000000001101")] = account

    os.RemoveAll(filepath.Join(s.contractPath, "temp"))
}

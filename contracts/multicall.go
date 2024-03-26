package contracts

import (
    "github.com/safe/SAFE4-genesis-tool/common"
    "github.com/safe/SAFE4-genesis-tool/types"
    "github.com/safe/SAFE4-genesis-tool/utils"
    "math/big"
    "os"
    "path/filepath"
)

type MulticallStorage struct {
    solcPath     string
    contractPath string
}

func NewMulticallStorage(tool *types.Tool) *MulticallStorage {
    return &MulticallStorage{
        solcPath:     tool.GetSolcPath(),
        contractPath: tool.GetContractPath(),
    }
}

func (s *MulticallStorage) Generate(alloc *types.GenesisAlloc) {
    utils.Compile(s.solcPath, s.contractPath, "Multicall.sol")

    codePath := filepath.Join(s.contractPath, "temp", "Multicall.bin-runtime")
    code, err := os.ReadFile(codePath)
    if err != nil {
        panic(err)
    }

    account := types.GenesisAccount{
        Balance: big.NewInt(0).String(),
        Code:    "0x" + string(code),
    }
    (*alloc)[common.HexToAddress("0x0000000000000000000000000000000000001100")] = account

    os.RemoveAll(filepath.Join(s.contractPath, "temp"))
}

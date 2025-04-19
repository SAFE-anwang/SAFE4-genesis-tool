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

    account.Storage = make(map[common.Hash]common.Hash)

    // name
    s.buildName(&account)

    // symbol
    s.buildSymbol(&account)

    // decimal
    s.buildDecimal(&account)

    (*alloc)[common.HexToAddress("0x0000000000000000000000000000000000001101")] = account

    os.RemoveAll(filepath.Join(s.contractPath, "temp"))
}

func (s *WSafeStorage) buildName(account *types.GenesisAccount) {
    curKey := big.NewInt(0)
    storageKeys, storageValues := utils.GetStorage4String(curKey, "Wrapped SAFE")
    for i := range storageKeys {
        account.Storage[storageKeys[i]] = storageValues[i]
    }
}

func (s *WSafeStorage) buildSymbol(account *types.GenesisAccount) {
    curKey := big.NewInt(1)
    storageKeys, storageValues := utils.GetStorage4String(curKey, "WSAFE")
    for i := range storageKeys {
        account.Storage[storageKeys[i]] = storageValues[i]
    }
}

func (s *WSafeStorage) buildDecimal(account *types.GenesisAccount) {
    curKey := big.NewInt(2)
    storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(18))
    account.Storage[storageKey] = storageValue
}

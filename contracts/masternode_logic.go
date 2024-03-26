package contracts

import (
    "github.com/safe/SAFE4-genesis-tool/common"
    "github.com/safe/SAFE4-genesis-tool/types"
    "github.com/safe/SAFE4-genesis-tool/utils"
    "math/big"
    "os"
    "path/filepath"
)

type MasterNodeLogicStorage struct {
    solcPath     string
    contractPath string
    ownerAddr    string
}

func NewMasterNodeLogicStorage(tool *types.Tool) *MasterNodeLogicStorage {
    return &MasterNodeLogicStorage{
        solcPath:     tool.GetSolcPath(),
        contractPath: tool.GetContractPath(),
        ownerAddr:    tool.GetOwnerAddress(),
    }
}

func (s *MasterNodeLogicStorage) Generate(alloc *types.GenesisAlloc) {
    utils.Compile(s.solcPath, s.contractPath, "MasterNodeLogic.sol")

    contractNames := [2]string{"TransparentUpgradeableProxy", "MasterNodeLogic"}
    contractAddrs := [2]string{"0x0000000000000000000000000000000000001025", "0x0000000000000000000000000000000000001026"}

    for i := range contractNames {
        code, err := os.ReadFile(filepath.Join(s.contractPath, "temp", contractNames[i]+".bin-runtime"))
        if err != nil {
            panic(err)
        }

        account := types.GenesisAccount{
            Balance: big.NewInt(0).String(),
            Code:    "0x" + string(code),
        }
        if contractNames[i] == "TransparentUpgradeableProxy" {
            account.Storage = make(map[common.Hash]common.Hash)
            account.Storage[common.BigToHash(big.NewInt(0))] = common.BigToHash(big.NewInt(1))
            account.Storage[common.BigToHash(big.NewInt(0x33))] = common.HexToHash(s.ownerAddr)
            account.Storage[common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")] = common.HexToHash(contractAddrs[1])
            account.Storage[common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")] = common.HexToHash(ProxyAdminAddr)
        }
        (*alloc)[common.HexToAddress(contractAddrs[i])] = account
    }

    os.RemoveAll(filepath.Join(s.contractPath, "temp"))
}

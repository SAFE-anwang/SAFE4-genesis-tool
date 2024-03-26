package contracts

import (
    "github.com/safe/SAFE4-genesis-tool/common"
    "github.com/safe/SAFE4-genesis-tool/types"
    "github.com/safe/SAFE4-genesis-tool/utils"
    "math/big"
    "os"
    "path/filepath"
)

var ProxyAdminAddr = "0x0000000000000000000000000000000000000999"

type ProxyAdminStorage struct {
    solcPath     string
    contractPath string
    ownerAddr    string
}

func NewProxyAdminStorage(tool *types.Tool) *ProxyAdminStorage {
    return &ProxyAdminStorage{
        solcPath:     tool.GetSolcPath(),
        contractPath: tool.GetContractPath(),
        ownerAddr:    tool.GetOwnerAddress(),
    }
}

func (s *ProxyAdminStorage) Generate(alloc *types.GenesisAlloc) {
    utils.Compile(s.solcPath, s.contractPath, "3rd/OpenZeppelin/openzeppelin-contracts/contracts/proxy/transparent/ProxyAdmin.sol")

    codePath := filepath.Join(s.contractPath, "temp", "ProxyAdmin.bin-runtime")
    code, err := os.ReadFile(codePath)
    if err != nil {
        panic(err)
    }

    account := types.GenesisAccount{
        Balance: big.NewInt(0).String(),
        Code:    "0x" + string(code),
    }
    account.Storage = make(map[common.Hash]common.Hash)
    account.Storage[common.BigToHash(big.NewInt(0))] = common.HexToHash(s.ownerAddr)
    (*alloc)[common.HexToAddress(ProxyAdminAddr)] = account

    os.RemoveAll(filepath.Join(s.contractPath, "temp"))
}

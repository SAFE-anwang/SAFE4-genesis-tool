package contracts

import (
	"github.com/safe/SAFE4-genesis-tool/common"
	"github.com/safe/SAFE4-genesis-tool/core"
	"github.com/safe/SAFE4-genesis-tool/utils"
	"math/big"
	"os"
	"path/filepath"
)

var ProxyAdminAddr = "0x0000000000000000000000000000000000000999"

type ProxyAdminStorage struct {
	workPath  string
	ownerAddr string
}

func NewProxyAdminStorage(workPath string, ownerAddr string) *ProxyAdminStorage {
	return &ProxyAdminStorage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *ProxyAdminStorage) Generate(alloc *core.GenesisAlloc) {
	utils.Compile(storage.workPath, "3rd/OpenZeppelin/openzeppelin-contracts/contracts/proxy/transparent/ProxyAdmin.sol")

	codePath := storage.workPath + "temp" + string(filepath.Separator) + "ProxyAdmin.bin-runtime"
	code, err := os.ReadFile(codePath)
	if err != nil {
		panic(err)
	}

	account := core.GenesisAccount{
		Balance: big.NewInt(0).String(),
		Code:    "0x" + string(code),
	}
	account.Storage = make(map[common.Hash]common.Hash)
	account.Storage[common.BigToHash(big.NewInt(0))] = common.HexToHash(storage.ownerAddr)
	(*alloc)[common.HexToAddress(ProxyAdminAddr)] = account
	os.RemoveAll(storage.workPath + "temp")
}

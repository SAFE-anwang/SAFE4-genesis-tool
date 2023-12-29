package contracts

import (
	"github.com/safe/SAFE4-genesis-tool/common"
	"github.com/safe/SAFE4-genesis-tool/core"
	"github.com/safe/SAFE4-genesis-tool/utils"
	"math/big"
	"os"
	"path/filepath"
)

type SuperNodeLogicStorage struct {
	workPath  string
	ownerAddr string
}

func NewSuperNodeLogicStorage(workPath string, ownerAddr string) *SuperNodeLogicStorage {
	return &SuperNodeLogicStorage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *SuperNodeLogicStorage) Generate(alloc *core.GenesisAlloc) {
	utils.Compile(storage.workPath, "SuperNodeLogic.sol")

	contractNames := [2]string{"TransparentUpgradeableProxy", "SuperNodeLogic"}
	contractAddrs := [2]string{"0x0000000000000000000000000000000000001035", "0x0000000000000000000000000000000000001036"}

	for i := range contractNames {
		codePath := storage.workPath + "temp" + string(filepath.Separator) + contractNames[i] + ".bin-runtime"
		code, err := os.ReadFile(codePath)
		if err != nil {
			panic(err)
		}

		account := core.GenesisAccount{
			Balance: big.NewInt(0).String(),
			Code:    string(code),
		}
		if contractNames[i] == "TransparentUpgradeableProxy" {
			account.Storage = make(map[common.Hash]common.Hash)
			account.Storage[common.BigToHash(big.NewInt(0))] = common.BigToHash(big.NewInt(1))
			account.Storage[common.BigToHash(big.NewInt(0x33))] = common.HexToHash(storage.ownerAddr)
			account.Storage[common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")] = common.HexToHash(contractAddrs[1])
			account.Storage[common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")] = common.HexToHash(ProxyAdminAddr)
		}
		(*alloc)[common.HexToAddress(contractAddrs[i])] = account
	}
	os.RemoveAll(storage.workPath + "temp")
}

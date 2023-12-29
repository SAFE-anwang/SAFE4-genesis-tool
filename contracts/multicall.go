package contracts

import (
	"github.com/safe/SAFE4-genesis-tool/common"
	"github.com/safe/SAFE4-genesis-tool/core"
	"github.com/safe/SAFE4-genesis-tool/utils"
	"math/big"
	"os"
	"path/filepath"
)

type MulticallStorage struct {
	workPath  string
	ownerAddr string
}

func NewMulticallStorage(workPath string, ownerAddr string) *MulticallStorage {
	return &MulticallStorage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *MulticallStorage) Generate(alloc *core.GenesisAlloc) {
	utils.Compile(storage.workPath, "Multicall.sol")

	codePath := storage.workPath + "temp" + string(filepath.Separator) + "Multicall.bin-runtime"
	code, err := os.ReadFile(codePath)
	if err != nil {
		panic(err)
	}

	account := core.GenesisAccount{
		Balance: big.NewInt(0).String(),
		Code:    string(code),
	}
	(*alloc)[common.HexToAddress("0x0000000000000000000000000000000000001100")] = account
	os.RemoveAll(storage.workPath + "temp")
}

package contracts

import (
	"encoding/hex"
	"github.com/safe/SAFE4-genesis-tool/common"
	"github.com/safe/SAFE4-genesis-tool/core"
	"github.com/safe/SAFE4-genesis-tool/utils"
	"math/big"
	"os"
	"path/filepath"
)

type MulticallStorage struct {
	workPath  string
	ownerAddr common.Address
}

func NewMulticallStorage(workPath string, ownerAddr common.Address) *MulticallStorage {
	return &MulticallStorage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *MulticallStorage) Generate(genesis *core.Genesis, allocAccounts *[]common.Address, mapAllocAccountStorageKeys *map[common.Address][]common.Hash) {
	utils.Compile(storage.workPath, "Multicall.sol")

	codePath := storage.workPath + "temp" + string(filepath.Separator) + "Multicall.bin-runtime"
	code, err := os.ReadFile(codePath)
	if err != nil {
		panic(err)
	}
	bs, err := hex.DecodeString(string(code))
	if err != nil {
		panic(err)
	}

	addr := common.HexToAddress("0x0000000000000000000000000000000000001100")
	*allocAccounts = append(*allocAccounts, addr)

	account := core.GenesisAccount{
		Balance: big.NewInt(0),
		Code:    bs,
	}
	genesis.Alloc[addr] = account

	var allocAccountStorageKeys []common.Hash
	allocAccountStorageKeys = append(allocAccountStorageKeys, common.BigToHash(big.NewInt(0)))
	(*mapAllocAccountStorageKeys)[addr] = allocAccountStorageKeys

	os.RemoveAll(storage.workPath + "temp")
}

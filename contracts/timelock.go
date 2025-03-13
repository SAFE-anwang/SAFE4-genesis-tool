package contracts

import (
	"github.com/safe/SAFE4-genesis-tool/common"
	"github.com/safe/SAFE4-genesis-tool/types"
	"github.com/safe/SAFE4-genesis-tool/utils"
	"math/big"
	"os"
	"path/filepath"
)

type TimeLockStorage struct {
	solcPath     string
	contractPath string
}

func NewTimeLockStorage(tool *types.Tool) *TimeLockStorage {
	return &TimeLockStorage{
		solcPath:     tool.GetSolcPath(),
		contractPath: tool.GetContractPath(),
	}
}

func (s *TimeLockStorage) Generate(alloc *types.GenesisAlloc) {
	utils.Compile(s.solcPath, s.contractPath, "TimeLock.sol")

	codePath := filepath.Join(s.contractPath, "temp", "TimeLock.bin-runtime")
	code, err := os.ReadFile(codePath)
	if err != nil {
		panic(err)
	}

	account := types.GenesisAccount{
		Balance: big.NewInt(0).String(),
		Code:    "0x" + string(code),
	}

	account.Storage = make(map[common.Hash]common.Hash)

	// minDelay
	s.buildMinDelay(&account)

	(*alloc)[common.HexToAddress("0x0000000000000000000000000000000000001103")] = account

	os.RemoveAll(filepath.Join(s.contractPath, "temp"))
}

func (s *TimeLockStorage) buildMinDelay(account *types.GenesisAccount) {
	curKey := big.NewInt(0)
	storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(3600))
	account.Storage[storageKey] = storageValue
}

package contracts

import (
	"github.com/safe/SAFE4-genesis-tool/common"
	"github.com/safe/SAFE4-genesis-tool/types"
	"github.com/safe/SAFE4-genesis-tool/utils"
	"math/big"
	"os"
	"path/filepath"
)

type MultiSigStorage struct {
	solcPath     string
	contractPath string
	owners       []string
}

func NewMultiSigStorage(tool *types.Tool) *MultiSigStorage {
	return &MultiSigStorage{
		solcPath:     tool.GetSolcPath(),
		contractPath: tool.GetContractPath(),
		owners:       tool.GetMultiSigOwners(),
	}
}

func (s *MultiSigStorage) Generate(alloc *types.GenesisAlloc) {
	utils.Compile(s.solcPath, s.contractPath, "MultisigWallet.sol")

	codePath := filepath.Join(s.contractPath, "temp", "MultisigWallet.bin-runtime")
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

	// required
	s.buildRequired(&account)

	// owners
	s.buildOwners(&account)

	// isOwner
	s.buildIsOwner(&account)

	(*alloc)[common.HexToAddress("0x0000000000000000000000000000000000001102")] = account

	os.RemoveAll(filepath.Join(s.contractPath, "temp"))
}

func (s *MultiSigStorage) buildMinDelay(account *types.GenesisAccount) {
	curKey := big.NewInt(0)
	storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(600))
	account.Storage[storageKey] = storageValue
}

func (s *MultiSigStorage) buildRequired(account *types.GenesisAccount) {
	curKey := big.NewInt(1)
	storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(3))
	account.Storage[storageKey] = storageValue
}

func (s *MultiSigStorage) buildOwners(account *types.GenesisAccount) {
	storageKey := common.BigToHash(big.NewInt(2))
	storageValue := common.BigToHash(big.NewInt(int64(len(s.owners))))
	account.Storage[storageKey] = storageValue

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(2))
	for i, owner := range s.owners {
		curKey := big.NewInt(0).Add(subKey, big.NewInt(int64(i)))
		subStorageKey, subStorageValue := utils.GetStorage4Addr(curKey, common.HexToAddress(owner))
		account.Storage[subStorageKey] = subStorageValue
	}
}

func (s *MultiSigStorage) buildIsOwner(account *types.GenesisAccount) {
	for _, owner := range s.owners {
		curKey := big.NewInt(0).SetBytes(utils.Keccak256_uint_address(3, common.HexToAddress(owner)))
		storageKey, storageValue := utils.GetStorage4Bool(curKey, true)
		account.Storage[storageKey] = storageValue
	}
}

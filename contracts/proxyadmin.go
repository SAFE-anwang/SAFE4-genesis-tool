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

var ProxyAdminAddr = common.HexToAddress("0x0000000000000000000000000000000000000999")

type ProxyAdminStorage struct {
	workPath  string
	ownerAddr common.Address
}

func NewProxyAdminStorage(workPath string, ownerAddr common.Address) *ProxyAdminStorage {
	return &ProxyAdminStorage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *ProxyAdminStorage) Generate(genesis *core.Genesis, allocAccounts *[]common.Address, mapAllocAccountStorageKeys *map[common.Address][]common.Hash) {
	utils.Compile(storage.workPath, "3rd/OpenZeppelin/openzeppelin-contracts/contracts/proxy/transparent/ProxyAdmin.sol")

	codePath := storage.workPath + "temp" + string(filepath.Separator) + "ProxyAdmin.bin-runtime"
	code, err := os.ReadFile(codePath)
	if err != nil {
		panic(err)
	}
	bs, err := hex.DecodeString(string(code))
	if err != nil {
		panic(err)
	}

	*allocAccounts = append(*allocAccounts, ProxyAdminAddr)

	account := core.GenesisAccount{
		Balance: big.NewInt(0),
		Code:    bs,
	}
	account.Storage = make(map[common.Hash]common.Hash)
	account.Storage[common.BigToHash(big.NewInt(0))] = common.HexToHash(storage.ownerAddr.Hex())
	genesis.Alloc[ProxyAdminAddr] = account

	var allocAccountStorageKeys []common.Hash
	allocAccountStorageKeys = append(allocAccountStorageKeys, common.BigToHash(big.NewInt(0)))
	(*mapAllocAccountStorageKeys)[ProxyAdminAddr] = allocAccountStorageKeys

	os.RemoveAll(storage.workPath + "temp")
}

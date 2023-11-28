package contracts

import (
	"encoding/hex"
	"encoding/json"
	"github.com/safe/SAFE4-genesis-tool/common"
	"github.com/safe/SAFE4-genesis-tool/core"
	"github.com/safe/SAFE4-genesis-tool/core/types"
	"github.com/safe/SAFE4-genesis-tool/utils"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
)

type PropertyStorage struct {
	workPath  string
	ownerAddr common.Address
}

func NewPropertyStorage(workPath string, ownerAddr common.Address) *PropertyStorage {
	return &PropertyStorage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *PropertyStorage) Generate(genesis *core.Genesis, allocAccounts *[]common.Address, mapAllocAccountStorageKeys *map[common.Address][]common.Hash) {
	utils.Compile(storage.workPath, "Property.sol")

	properties := storage.load()

	contractNames := [2]string{"TransparentUpgradeableProxy", "Property"}
	contractAddrs := [2]string{"0x0000000000000000000000000000000000001000", "0x0000000000000000000000000000000000001001"}

	for i := range contractNames {
		key := contractNames[i]
		value := contractAddrs[i]

		codePath := storage.workPath + "temp" + string(filepath.Separator) + key + ".bin-runtime"
		code, err := os.ReadFile(codePath)
		if err != nil {
			panic(err)
		}

		bs, err := hex.DecodeString(string(code))
		if err != nil {
			panic(err)
		}

		addr := common.HexToAddress(value)
		*allocAccounts = append(*allocAccounts, addr)

		account := core.GenesisAccount{
			Balance: big.NewInt(0),
			Code:    bs,
		}
		var allocAccountStorageKeys []common.Hash
		if key == "TransparentUpgradeableProxy" {
			account.Storage = make(map[common.Hash]common.Hash)

			account.Storage[common.BigToHash(big.NewInt(0))] = common.BigToHash(big.NewInt(1))
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.BigToHash(big.NewInt(0)))

			account.Storage[common.BigToHash(big.NewInt(0x33))] = common.HexToHash(storage.ownerAddr.Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.BigToHash(big.NewInt(0x33)))

			account.Storage[common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")] = common.HexToHash(common.HexToAddress(contractAddrs[0]).Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc"))

			account.Storage[common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")] = common.HexToHash(ProxyAdminAddr.Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103"))

			// properties
			storage.buildProperties(&account, &allocAccountStorageKeys, properties)

			// confirmedNames
			storage.buildConfirmedNames(&account, &allocAccountStorageKeys, properties)
		}

		if len(allocAccountStorageKeys) != 0 {
			(*mapAllocAccountStorageKeys)[addr] = allocAccountStorageKeys
		}

		genesis.Alloc[addr] = account
	}
	os.RemoveAll(storage.workPath + "temp")
}

func (storage *PropertyStorage) load() *[]types.PropertyInfo {
	jsonFile, err := os.Open(storage.workPath + utils.GetDataDir() + string(filepath.Separator) + "Property.info")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	properties := new([]types.PropertyInfo)
	err = json.Unmarshal(jsonData, properties)
	if err != nil {
		panic(err)
	}
	return properties
}

func (storage *PropertyStorage) buildProperties(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, properties *[]types.PropertyInfo) {
	var curKey *big.Int
	for _, property := range *properties {
		storage.calcName(account, allocAccountStorageKeys, property, &curKey)
		storage.calcValue(account, allocAccountStorageKeys, property, &curKey)
		storage.calcDescription(account, allocAccountStorageKeys, property, &curKey)
		storage.calcCreateHeight(account, allocAccountStorageKeys, property, &curKey)
		storage.calcUpdateHeight(account, allocAccountStorageKeys, property, &curKey)
	}
}

func (storage *PropertyStorage) calcName(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, property types.PropertyInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_string(101, property.Name))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, property.Name)
	if len(storageKeys) != len(storageValues) {
		panic("get storage failed")
	}
	for i := range storageKeys {
		account.Storage[storageKeys[i]] = storageValues[i]
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKeys[i])
	}
}

func (storage *PropertyStorage) calcValue(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, property types.PropertyInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, property.Value)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *PropertyStorage) calcDescription(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, property types.PropertyInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, property.Description)
	if len(storageKeys) != len(storageValues) {
		panic("get storage failed")
	}
	for i := range storageKeys {
		account.Storage[storageKeys[i]] = storageValues[i]
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKeys[i])
	}
}

func (storage *PropertyStorage) calcCreateHeight(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, property types.PropertyInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, property.CreateHeight)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *PropertyStorage) calcUpdateHeight(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, property types.PropertyInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, property.UpdateHeight)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *PropertyStorage) buildConfirmedNames(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, properties *[]types.PropertyInfo) {
	storageKey := common.BigToHash(big.NewInt(102))
	storageValue := common.BigToHash(big.NewInt(int64(len(*properties))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(102))
	for i, property := range *properties {
		curKey := big.NewInt(0).Add(subKey, big.NewInt(int64(i)))
		subStorageKeys, subStorageValues := utils.GetStorage4String(curKey, property.Name)
		if len(subStorageKeys) != len(subStorageValues) {
			panic("get storage failed")
		}
		for k := range subStorageKeys {
			account.Storage[subStorageKeys[k]] = subStorageValues[k]
			*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKeys[k])
		}
	}
}

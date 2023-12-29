package contracts

import (
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
	ownerAddr string
}

func NewPropertyStorage(workPath string, ownerAddr string) *PropertyStorage {
	return &PropertyStorage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *PropertyStorage) Generate(alloc *core.GenesisAlloc) {
	utils.Compile(storage.workPath, "Property.sol")

	properties := storage.load()

	contractNames := [2]string{"TransparentUpgradeableProxy", "Property"}
	contractAddrs := [2]string{"0x0000000000000000000000000000000000001000", "0x0000000000000000000000000000000000001001"}

	for i := range contractNames {
		key := contractNames[i]

		codePath := storage.workPath + "temp" + string(filepath.Separator) + key + ".bin-runtime"
		code, err := os.ReadFile(codePath)
		if err != nil {
			panic(err)
		}

		account := core.GenesisAccount{
			Balance: big.NewInt(0).String(),
			Code:    "0x" + string(code),
		}
		if key == "TransparentUpgradeableProxy" {
			account.Storage = make(map[common.Hash]common.Hash)
			account.Storage[common.BigToHash(big.NewInt(0))] = common.BigToHash(big.NewInt(1))
			account.Storage[common.BigToHash(big.NewInt(0x33))] = common.HexToHash(storage.ownerAddr)
			account.Storage[common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")] = common.HexToHash(contractAddrs[1])
			account.Storage[common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")] = common.HexToHash(ProxyAdminAddr)

			// properties
			storage.buildProperties(&account, properties)

			// confirmedNames
			storage.buildConfirmedNames(&account, properties)
		}
		(*alloc)[common.HexToAddress(contractAddrs[i])] = account
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

func (storage *PropertyStorage) buildProperties(account *core.GenesisAccount, properties *[]types.PropertyInfo) {
	var curKey *big.Int
	for _, property := range *properties {
		storage.calcName(account, property, &curKey)
		storage.calcValue(account, property, &curKey)
		storage.calcDescription(account, property, &curKey)
	}
}

func (storage *PropertyStorage) calcName(account *core.GenesisAccount, property types.PropertyInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_string(101, property.Name))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, property.Name)
	for i := range storageKeys {
		account.Storage[storageKeys[i]] = storageValues[i]
	}
}

func (storage *PropertyStorage) calcValue(account *core.GenesisAccount, property types.PropertyInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, property.Value)
	account.Storage[storageKey] = storageValue
}

func (storage *PropertyStorage) calcDescription(account *core.GenesisAccount, property types.PropertyInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, property.Description)
	for i := range storageKeys {
		account.Storage[storageKeys[i]] = storageValues[i]
	}
}

func (storage *PropertyStorage) calcCreateHeight(account *core.GenesisAccount, property types.PropertyInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, property.CreateHeight)
	account.Storage[storageKey] = storageValue
}

func (storage *PropertyStorage) calcUpdateHeight(account *core.GenesisAccount, property types.PropertyInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, property.UpdateHeight)
	account.Storage[storageKey] = storageValue
}

func (storage *PropertyStorage) buildConfirmedNames(account *core.GenesisAccount, properties *[]types.PropertyInfo) {
	storageKey := common.BigToHash(big.NewInt(102))
	storageValue := common.BigToHash(big.NewInt(int64(len(*properties))))
	account.Storage[storageKey] = storageValue

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(102))
	for i, property := range *properties {
		curKey := big.NewInt(0).Add(subKey, big.NewInt(int64(i)))
		subStorageKeys, subStorageValues := utils.GetStorage4String(curKey, property.Name)
		for k := range subStorageKeys {
			account.Storage[subStorageKeys[k]] = subStorageValues[k]
		}
	}
}

package contracts

import (
    "encoding/json"
    "github.com/safe/SAFE4-genesis-tool/common"
    "github.com/safe/SAFE4-genesis-tool/types"
    "github.com/safe/SAFE4-genesis-tool/utils"
    "io/ioutil"
    "math/big"
    "os"
    "path/filepath"
)

type PropertyStorage struct {
    dataPath     string
    solcPath     string
    contractPath string
    ownerAddr    string
}

func NewPropertyStorage(tool *types.Tool) *PropertyStorage {
    return &PropertyStorage{
        dataPath:     tool.GetDataPath(),
        solcPath:     tool.GetSolcPath(),
        contractPath: tool.GetContractPath(),
        ownerAddr:    tool.GetOwnerAddress(),
    }
}

func (s *PropertyStorage) Generate(alloc *types.GenesisAlloc) {
    utils.Compile(s.solcPath, s.contractPath, "Property.sol")

    properties := s.load()

    contractNames := [2]string{"TransparentUpgradeableProxy", "Property"}
    contractAddrs := [2]string{"0x0000000000000000000000000000000000001000", "0x0000000000000000000000000000000000001001"}

    for i := range contractNames {
        codePath := filepath.Join(s.contractPath, "temp", contractNames[i]+".bin-runtime")
        code, err := os.ReadFile(codePath)
        if err != nil {
            panic(err)
        }

        account := types.GenesisAccount{
            Balance: big.NewInt(0).String(),
            Code:    "0x" + string(code),
        }
        if contractNames[i] == "TransparentUpgradeableProxy" {
            account.Storage = make(map[common.Hash]common.Hash)
            account.Storage[common.BigToHash(big.NewInt(0))] = common.BigToHash(big.NewInt(1))
            account.Storage[common.BigToHash(big.NewInt(0x33))] = common.HexToHash(s.ownerAddr)
            account.Storage[common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")] = common.HexToHash(contractAddrs[1])
            account.Storage[common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")] = common.HexToHash(ProxyAdminAddr)

            // properties
            s.buildProperties(&account, properties)

            // confirmedNames
            s.buildConfirmedNames(&account, properties)
        }
        (*alloc)[common.HexToAddress(contractAddrs[i])] = account
    }

    os.RemoveAll(filepath.Join(s.contractPath, "temp"))
}

func (s *PropertyStorage) load() *[]types.PropertyInfo {
    jsonFile, err := os.Open(filepath.Join(s.dataPath, "Property.info"))
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

func (s *PropertyStorage) buildProperties(account *types.GenesisAccount, properties *[]types.PropertyInfo) {
    var curKey *big.Int
    for _, property := range *properties {
        s.calcName(account, property, &curKey)
        s.calcValue(account, property, &curKey)
        s.calcDescription(account, property, &curKey)
    }
}

func (s *PropertyStorage) calcName(account *types.GenesisAccount, property types.PropertyInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_string(101, property.Name))
    storageKeys, storageValues := utils.GetStorage4String(*curKey, property.Name)
    for i := range storageKeys {
        account.Storage[storageKeys[i]] = storageValues[i]
    }
}

func (s *PropertyStorage) calcValue(account *types.GenesisAccount, property types.PropertyInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue := utils.GetStorage4Int(*curKey, property.Value)
    account.Storage[storageKey] = storageValue
}

func (s *PropertyStorage) calcDescription(account *types.GenesisAccount, property types.PropertyInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKeys, storageValues := utils.GetStorage4String(*curKey, property.Description)
    for i := range storageKeys {
        account.Storage[storageKeys[i]] = storageValues[i]
    }
}

func (s *PropertyStorage) calcCreateHeight(account *types.GenesisAccount, property types.PropertyInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue := utils.GetStorage4Int(*curKey, property.CreateHeight)
    account.Storage[storageKey] = storageValue
}

func (s *PropertyStorage) calcUpdateHeight(account *types.GenesisAccount, property types.PropertyInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue := utils.GetStorage4Int(*curKey, property.UpdateHeight)
    account.Storage[storageKey] = storageValue
}

func (s *PropertyStorage) buildConfirmedNames(account *types.GenesisAccount, properties *[]types.PropertyInfo) {
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

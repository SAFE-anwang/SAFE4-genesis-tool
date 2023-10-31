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

type SuperNodeStateStorage struct {
	workPath string
	ownerAddr common.Address
}

func NewSuperNodeStateStorage(workPath string, ownerAddr common.Address) *SuperNodeStateStorage {
	return &SuperNodeStateStorage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *SuperNodeStateStorage) Generate(genesis *core.Genesis, allocAccounts *[]common.Address, mapAllocAccountStorageKeys *map[common.Address][]common.Hash) {
	utils.Compile(storage.workPath, "SuperNodeState.sol")

	supernodes := storage.LoadSuperNode()

	contractNames := [3]string{"SuperNodeState", "ProxyAdmin", "TransparentUpgradeableProxy"}
	contractAddrs := [3]string{"0x0000000000000000000000000000000000001060", "0x0000000000000000000000000000000000001061", "0x0000000000000000000000000000000000001062"}

	for i, _ := range contractNames {
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

		account := core.GenesisAccount{
			Balance: big.NewInt(0),
			Code: bs,
		}
		addr := common.HexToAddress(value)
		*allocAccounts = append(*allocAccounts, addr)
		var allocAccountStorageKeys []common.Hash
		if key == "ProxyAdmin" {
			account.Storage = make(map[common.Hash]common.Hash)
			account.Storage[common.BigToHash(big.NewInt(0))] = common.HexToHash(storage.ownerAddr.Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.BigToHash(big.NewInt(0)))
		} else if key == "TransparentUpgradeableProxy" {
			account.Storage = make(map[common.Hash]common.Hash)

			account.Storage[common.BigToHash(big.NewInt(0))] = common.BigToHash(big.NewInt(1))
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.BigToHash(big.NewInt(0)))

			account.Storage[common.BigToHash(big.NewInt(0x33))] = common.HexToHash(storage.ownerAddr.Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.BigToHash(big.NewInt(0x33)))

			account.Storage[common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")] = common.HexToHash(common.HexToAddress(contractAddrs[0]).Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc"))
			
			account.Storage[common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")] = common.HexToHash(common.HexToAddress(contractAddrs[1]).Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103"))

			// ids
			storage.buildIDs(&account, &allocAccountStorageKeys, supernodes)

			// id2index
			storage.buildId2Index(&account, &allocAccountStorageKeys, supernodes)

			// id2state
			storage.buildId2State(&account, &allocAccountStorageKeys, supernodes)

			// id2entries
			//storage.buildId2Entries(&account, &allocAccountStorageKeys, supernodes)
		}

		if len(allocAccountStorageKeys) != 0 {
			(*mapAllocAccountStorageKeys)[addr] = allocAccountStorageKeys
		}

		genesis.Alloc[addr] = account
	}
	os.RemoveAll(storage.workPath + "temp")
}

func (storage *SuperNodeStateStorage) LoadSuperNode() *[]types.SuperNodeInfo {
	jsonFile, err := os.Open(storage.workPath + "data" + string(filepath.Separator) + "SuperNode.info")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	supernodes := new([]types.SuperNodeInfo)
	err = json.Unmarshal(jsonData, supernodes)
	if err != nil {
		panic(err)
	}
	return supernodes
}

func (storage *SuperNodeStateStorage) buildIDs(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernodes *[]types.SuperNodeInfo) {
	storageKey := common.BigToHash(big.NewInt(101))
	storageValue := common.BigToHash(big.NewInt(int64(len(*supernodes))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(101))
	for i, supernode := range *supernodes {
		curKey := big.NewInt(0).Add(subKey, big.NewInt(int64(i)))
		subStorageKey, subStorageValue := utils.GetStorage4Int(curKey, supernode.Id)
		account.Storage[subStorageKey] = subStorageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKey)
	}
}

func (storage *SuperNodeStateStorage) buildId2Index(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernodes *[]types.SuperNodeInfo) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for i, sn := range *supernodes {
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(102, sn.Id.Int64()))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(int64(i + 1)))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
	}
}

func (storage *SuperNodeStateStorage) buildId2State(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernodes *[]types.SuperNodeInfo) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for _, sn := range *supernodes {
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(103, sn.Id.Int64()))
		storageKey, storageValue = utils.GetStorage4Int(curKey, sn.StateInfo.State)
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
	}
}
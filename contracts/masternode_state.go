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

type MasterNodeStateStorage struct {
	workPath string
	ownerAddr common.Address
}

func NewMasterNodeStateStorage(workPath string, ownerAddr common.Address) *MasterNodeStateStorage {
	return &MasterNodeStateStorage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *MasterNodeStateStorage) Generate(genesis *core.Genesis, allocAccounts *[]common.Address, mapAllocAccountStorageKeys *map[common.Address][]common.Hash) {
	utils.Compile(storage.workPath, "MasterNodeState.sol")

	masternodes := storage.LoadMasterNode()

	contractNames := [3]string{"MasterNodeState", "ProxyAdmin", "TransparentUpgradeableProxy"}
	contractAddrs := [3]string{"0x0000000000000000000000000000000000001050", "0x0000000000000000000000000000000000001051", "0x0000000000000000000000000000000000001052"}

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
			storage.buildIDs(&account, &allocAccountStorageKeys, masternodes)

			// id2index
			storage.buildId2Index(&account, &allocAccountStorageKeys, masternodes)

			// id2state
			storage.buildId2State(&account, &allocAccountStorageKeys, masternodes)

			// id2entries
			//storage.buildId2Entries(&account, &allocAccountStorageKeys, masternodes)
		}

		if len(allocAccountStorageKeys) != 0 {
			(*mapAllocAccountStorageKeys)[addr] = allocAccountStorageKeys
		}

		genesis.Alloc[addr] = account
	}
	os.RemoveAll(storage.workPath + "temp")
}

func (storage *MasterNodeStateStorage) LoadMasterNode() *[]types.MasterNodeInfo {
	jsonFile, err := os.Open(storage.workPath + "data" + string(filepath.Separator) + "MasterNode.info")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	masternodes := new([]types.MasterNodeInfo)
	err = json.Unmarshal(jsonData, masternodes)
	if err != nil {
		panic(err)
	}
	return masternodes
}

func (storage *MasterNodeStateStorage) buildIDs(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternodes *[]types.MasterNodeInfo) {
	storageKey := common.BigToHash(big.NewInt(101))
	storageValue := common.BigToHash(big.NewInt(int64(len(*masternodes))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(101))
	for i, masternode := range *masternodes {
		curKey := big.NewInt(0).Add(subKey, big.NewInt(int64(i)))
		subStorageKey, subStorageValue := utils.GetStorage4Int(curKey, masternode.Id)
		account.Storage[subStorageKey] = subStorageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKey)
	}
}

func (storage *MasterNodeStateStorage) buildId2Index(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternodes *[]types.MasterNodeInfo) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for i, mn := range *masternodes {
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(102, mn.Id.Int64()))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(int64(i)))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
	}
}

func (storage *MasterNodeStateStorage) buildId2State(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternodes *[]types.MasterNodeInfo) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for _, mn := range *masternodes {
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(103, mn.Id.Int64()))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(1))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
	}
}
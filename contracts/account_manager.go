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

type AccountManagerStorage struct {
	workPath string
	ownerAddr common.Address
}

func NewAccountManagerStorage(workPath string, ownerAddr common.Address) *AccountManagerStorage {
	return &AccountManagerStorage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *AccountManagerStorage) Generate(genesis *core.Genesis, allocAccounts *[]common.Address, mapAllocAccountStorageKeys *map[common.Address][]common.Hash) {
	utils.Compile(storage.workPath, "AccountManager.sol")

	masternodes := storage.LoadMasterNode()
	supernodes := storage.LoadSuperNode()

	totalAmount := big.NewInt(0)
	var addrs []common.Address
	var amounts []*big.Int
	for _, masternode := range *masternodes {
		totalAmount = totalAmount.Add(totalAmount, masternode.Amount)
		addrs = append(addrs, masternode.Addr)
		amounts = append(amounts, masternode.Amount)
	}
	for _, supernode := range *supernodes {
		totalAmount = totalAmount.Add(totalAmount, supernode.Amount)
		addrs = append(addrs, supernode.Addr)
		amounts = append(amounts, supernode.Amount)
	}

	contractNames := [3]string{"AccountManager", "ProxyAdmin", "TransparentUpgradeableProxy"}
	contractAddrs := [3]string{"0x0000000000000000000000000000000000001010", "0x0000000000000000000000000000000000001011", "0x0000000000000000000000000000000000001012"}

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
			account.Balance = totalAmount
			account.Storage = make(map[common.Hash]common.Hash)

			account.Storage[common.BigToHash(big.NewInt(0))] = common.BigToHash(big.NewInt(1))
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.BigToHash(big.NewInt(0)))

			account.Storage[common.BigToHash(big.NewInt(0x33))] = common.HexToHash(storage.ownerAddr.Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.BigToHash(big.NewInt(0x33)))

			account.Storage[common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")] = common.HexToHash(common.HexToAddress(contractAddrs[0]).Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc"))
			
			account.Storage[common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")] = common.HexToHash(common.HexToAddress(contractAddrs[1]).Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103"))

			// record_no
			storage.buildRecordNo(&account, &allocAccountStorageKeys, len(addrs))

			// addr2records
			storage.buildAddr2records(&account, &allocAccountStorageKeys, addrs, amounts)

			// id2index
			storage.buildID2index(&account, &allocAccountStorageKeys, addrs)

			// id2addr
			storage.buildID2addr(&account, &allocAccountStorageKeys, addrs)

			// id2useinfo
			storage.buildID2useInfo(&account, &allocAccountStorageKeys, addrs)
		}

		if len(allocAccountStorageKeys) != 0 {
			(*mapAllocAccountStorageKeys)[addr] = allocAccountStorageKeys
		}

		genesis.Alloc[addr] = account
	}
	os.RemoveAll(storage.workPath + "temp")
}

func (storage *AccountManagerStorage) LoadMasterNode() *[]types.MasterNodeInfo {
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

func (storage *AccountManagerStorage) LoadSuperNode() *[]types.SuperNodeInfo {
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

func (storage *AccountManagerStorage) buildRecordNo(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, count int) {
	curKey := big.NewInt(102)
	storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(int64(count)))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *AccountManagerStorage) buildAddr2records(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, addrs []common.Address, amounts []*big.Int) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for i, addr := range addrs {
		// size
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_address(103, addr))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(1))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

		// id
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_bytes32(common.BigToHash(curKey).Hex()))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(int64(i + 1)))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
		// addr
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Addr(curKey, addr)
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
		// amount
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, amounts[i])
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
		// lockDay
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(720))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
		// startHeight
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(0))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
		// unlockHeight
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(720 * 24 * 3600 / 30))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
	}
}

func (storage *AccountManagerStorage) buildID2index(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, addrs []common.Address) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for i, _ := range addrs {
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(104, int64(i + 1)))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(0))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
	}
}

func (storage *AccountManagerStorage) buildID2addr(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, addrs []common.Address) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for i, addr := range addrs {
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(105, int64(i + 1)))
		storageKey, storageValue = utils.GetStorage4Addr(curKey, addr)
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
	}
}

func (storage *AccountManagerStorage) buildID2useInfo(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, addrs []common.Address) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for i, addr := range addrs {
		// specialAddr
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(105, int64(i + 1)))
		storageKey, storageValue = utils.GetStorage4Addr(curKey, addr)
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
		// freezeHeight
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(0))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
		// unfreezeHeight
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(720 * 24 * 3600 / 30))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
		// voterAddr
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(105, int64(i + 1)))
		storageKey, storageValue = utils.GetStorage4Addr(curKey, common.Address{})
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
		// voteHeight
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(0))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
		// releaseHeight
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(0))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
	}
}
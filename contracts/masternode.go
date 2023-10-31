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

type MasterNodeStorage struct {
	workPath string
	ownerAddr common.Address
}

func NewMasterNodeStorage(workPath string, ownerAddr common.Address) *MasterNodeStorage {
	return &MasterNodeStorage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *MasterNodeStorage) Generate(genesis *core.Genesis, allocAccounts *[]common.Address, mapAllocAccountStorageKeys *map[common.Address][]common.Hash) {
	utils.Compile(storage.workPath, "MasterNode.sol")

	masternodes := storage.load()

	contractNames := [3]string{"MasterNode", "ProxyAdmin", "TransparentUpgradeableProxy"}
	contractAddrs := [3]string{"0x0000000000000000000000000000000000001020", "0x0000000000000000000000000000000000001021", "0x0000000000000000000000000000000000001022"}

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

			// mn_no
			storage.buildMnNo(&account, &allocAccountStorageKeys, masternodes)

			// masternodes
			storage.buildMasterNodes(&account, &allocAccountStorageKeys, masternodes)

			// mnIDs
			storage.buildMnIDs(&account, &allocAccountStorageKeys, masternodes)

			// mnID2addr
			storage.buildMnID2addr(&account, &allocAccountStorageKeys, masternodes)

			// mnIP2addr
			storage.buildMnEnode2addr(&account, &allocAccountStorageKeys, masternodes)
		}

		if len(allocAccountStorageKeys) != 0 {
			(*mapAllocAccountStorageKeys)[addr] = allocAccountStorageKeys
		}

		genesis.Alloc[addr] = account
	}
	os.RemoveAll(storage.workPath + "temp")
}

func (storage *MasterNodeStorage) load() *[]types.MasterNodeInfo {
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

func (storage *MasterNodeStorage) buildMnNo(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternodes *[]types.MasterNodeInfo) {
	curKey := big.NewInt(101)
	storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(int64(len(*masternodes))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *MasterNodeStorage) buildMasterNodes(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternodes *[]types.MasterNodeInfo) {
	var curKey *big.Int
	for _, masternode := range *masternodes {
		storage.calcId(account, allocAccountStorageKeys, masternode, &curKey)
		storage.calcAddr(account, allocAccountStorageKeys, masternode, &curKey)
		storage.calcCreator(account, allocAccountStorageKeys, masternode, &curKey)
		storage.calcAmount(account, allocAccountStorageKeys, masternode, &curKey)
		storage.calcEnode(account, allocAccountStorageKeys, masternode, &curKey)
		storage.calcDesc(account, allocAccountStorageKeys, masternode, &curKey)
		storage.calcIsOfficial(account, allocAccountStorageKeys, masternode, &curKey)
		storage.calcStateInfo(account, allocAccountStorageKeys, masternode, &curKey)
		storage.calcFounders(account, allocAccountStorageKeys, masternode, &curKey)
		storage.calcIncentivePlan(account, allocAccountStorageKeys, masternode, &curKey)
		storage.calcLastRewardHeight(account, allocAccountStorageKeys, masternode, &curKey)
		storage.calcCreateHeight(account, allocAccountStorageKeys, masternode, &curKey)
		storage.calcUpdateHeight(account, allocAccountStorageKeys, masternode, &curKey)
	}
}

func (storage *MasterNodeStorage) calcId(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternode types.MasterNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_address(102, masternode.Addr))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, masternode.Id)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *MasterNodeStorage) calcAddr(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternode types.MasterNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Addr(*curKey, masternode.Addr)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *MasterNodeStorage) calcCreator(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternode types.MasterNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Addr(*curKey, masternode.Creator)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *MasterNodeStorage) calcAmount(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternode types.MasterNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, masternode.Amount)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *MasterNodeStorage) calcEnode(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternode types.MasterNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, masternode.Enode)
	if len(storageKeys) != len(storageValues) {
		panic("get storage failed")
	}
	for i, _ := range storageKeys {
		account.Storage[storageKeys[i]] = storageValues[i]
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKeys[i])
	}
}

func (storage *MasterNodeStorage) calcDesc(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternode types.MasterNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, masternode.Description)
	if len(storageKeys) != len(storageValues) {
		panic("get storage failed")
	}
	for i, _ := range storageKeys {
		account.Storage[storageKeys[i]] = storageValues[i]
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKeys[i])
	}
}

func (storage *MasterNodeStorage) calcIsOfficial(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternode types.MasterNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Bool(*curKey, masternode.IsOfficial)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *MasterNodeStorage) calcStateInfo(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternode types.MasterNodeInfo, curKey **big.Int) {
	var storageKey, storageValue common.Hash
	// state
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue = utils.GetStorage4Int(*curKey, masternode.StateInfo.State)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	// height
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue = utils.GetStorage4Int(*curKey, masternode.StateInfo.Height)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *MasterNodeStorage) calcFounders(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternode types.MasterNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey := common.BigToHash(*curKey)
	storageValue := common.BigToHash(big.NewInt(int64(len(masternode.Founders))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_bytes32(common.BigToHash(*curKey).Hex()))
	var subStorageKey, subStorageValue common.Hash
	for _, founder := range masternode.Founders {
		// lockID
		subStorageKey, subStorageValue = utils.GetStorage4Int(subKey, founder.LockID)
		account.Storage[subStorageKey] = subStorageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKey)
		subKey = subKey.Add(subKey, big.NewInt(1))

		// addr
		subStorageKey, subStorageValue = utils.GetStorage4Addr(subKey, founder.Addr)
		account.Storage[subStorageKey] = subStorageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKey)
		subKey = subKey.Add(subKey, big.NewInt(1))

		// amount
		subStorageKey, subStorageValue = utils.GetStorage4Int(subKey, founder.Amount)
		account.Storage[subStorageKey] = subStorageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKey)
		subKey = subKey.Add(subKey, big.NewInt(1))

		// height
		subStorageKey, subStorageValue = utils.GetStorage4Int(subKey, founder.Height)
		account.Storage[subStorageKey] = subStorageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKey)
		subKey = subKey.Add(subKey, big.NewInt(1))
	}
}

func (storage *MasterNodeStorage) calcIncentivePlan(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternode types.MasterNodeInfo, curKey **big.Int) {
	var storageKey, storageValue common.Hash
	// creator
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue = utils.GetStorage4Int(*curKey, masternode.IncentivePlan.Creator)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	// partner
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue = utils.GetStorage4Int(*curKey, masternode.IncentivePlan.Partner)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	// voter
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue = utils.GetStorage4Int(*curKey, masternode.IncentivePlan.Voter)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *MasterNodeStorage) calcLastRewardHeight(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternode types.MasterNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, masternode.LastRewardHeight)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *MasterNodeStorage) calcCreateHeight(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternode types.MasterNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, masternode.CreateHeight)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *MasterNodeStorage) calcUpdateHeight(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternode types.MasterNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, masternode.UpdateHeight)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *MasterNodeStorage) buildMnIDs(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternodes *[]types.MasterNodeInfo) {
	storageKey := common.BigToHash(big.NewInt(103))
	storageValue := common.BigToHash(big.NewInt(int64(len(*masternodes))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(103))
	for i, masternode := range *masternodes {
		curKey := big.NewInt(0).Add(subKey, big.NewInt(int64(i)))
		subStorageKey, subStorageValue := utils.GetStorage4Int(curKey, masternode.Id)
		account.Storage[subStorageKey] = subStorageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKey)
	}
}

func (storage *MasterNodeStorage) buildMnID2addr(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternodes *[]types.MasterNodeInfo) {
	for _, masternode := range *masternodes {
		curKey := big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(104, masternode.Id.Int64()))
		storageKey, storageValue := utils.GetStorage4Addr(curKey, masternode.Addr)
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
	}
}

func (storage *MasterNodeStorage) buildMnEnode2addr(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, masternodes *[]types.MasterNodeInfo) {
	for _, masternode := range *masternodes {
		curKey := big.NewInt(0).SetBytes(utils.Keccak256_uint_string(105, masternode.Enode))
		storageKey, storageValue := utils.GetStorage4Addr(curKey, masternode.Addr)
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
	}
}
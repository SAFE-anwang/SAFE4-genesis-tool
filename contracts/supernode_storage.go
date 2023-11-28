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

type SuperNodeStorageStorage struct {
	workPath  string
	ownerAddr common.Address
}

func NewSuperNodeStorageStorage(workPath string, ownerAddr common.Address) *SuperNodeStorageStorage {
	return &SuperNodeStorageStorage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *SuperNodeStorageStorage) Generate(genesis *core.Genesis, allocAccounts *[]common.Address, mapAllocAccountStorageKeys *map[common.Address][]common.Hash) {
	utils.Compile(storage.workPath, "SuperNodeStorage.sol")

	supernodes := storage.load()

	contractNames := [2]string{"TransparentUpgradeableProxy", "SuperNodeStorage"}
	contractAddrs := [2]string{"0x0000000000000000000000000000000000001030", "0x0000000000000000000000000000000000001031"}

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

			account.Storage[common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")] = common.HexToHash(common.HexToAddress(contractAddrs[1]).Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc"))

			account.Storage[common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")] = common.HexToHash(ProxyAdminAddr.Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103"))

			// no
			storage.buildNo(&account, &allocAccountStorageKeys, supernodes)

			// addr2info
			storage.buildAddr2Info(&account, &allocAccountStorageKeys, supernodes)

			// ids
			storage.buildIDs(&account, &allocAccountStorageKeys, supernodes)

			// id2addr
			storage.buildID2Addr(&account, &allocAccountStorageKeys, supernodes)

			// name2addr
			storage.buildName2Addr(&account, &allocAccountStorageKeys, supernodes)

			// enode2addr
			storage.buildEnode2Addr(&account, &allocAccountStorageKeys, supernodes)
		}

		if len(allocAccountStorageKeys) != 0 {
			(*mapAllocAccountStorageKeys)[addr] = allocAccountStorageKeys
		}

		genesis.Alloc[addr] = account
	}
	os.RemoveAll(storage.workPath + "temp")
}

func (storage *SuperNodeStorageStorage) load() *[]types.SuperNodeInfo {
	jsonFile, err := os.Open(storage.workPath + utils.GetDataDir() + string(filepath.Separator) + "SuperNode.info")
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

func (storage *SuperNodeStorageStorage) buildNo(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernodes *[]types.SuperNodeInfo) {
	curKey := big.NewInt(101)
	storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(int64(len(*supernodes))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *SuperNodeStorageStorage) buildAddr2Info(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernodes *[]types.SuperNodeInfo) {
	var curKey *big.Int
	for _, supernode := range *supernodes {
		storage.calcId(account, allocAccountStorageKeys, supernode, &curKey)
		storage.calcName(account, allocAccountStorageKeys, supernode, &curKey)
		storage.calcAddr(account, allocAccountStorageKeys, supernode, &curKey)
		storage.calcCreator(account, allocAccountStorageKeys, supernode, &curKey)
		storage.calcEnode(account, allocAccountStorageKeys, supernode, &curKey)
		storage.calcDesc(account, allocAccountStorageKeys, supernode, &curKey)
		storage.calcIsOfficial(account, allocAccountStorageKeys, supernode, &curKey)
		storage.calcStateInfo(account, allocAccountStorageKeys, supernode, &curKey)
		storage.calcFounders(account, allocAccountStorageKeys, supernode, &curKey)
		storage.calcIncentivePlan(account, allocAccountStorageKeys, supernode, &curKey)
		storage.calcVoteInfo(account, allocAccountStorageKeys, supernode, &curKey)
		storage.calcLastRewardHeight(account, allocAccountStorageKeys, supernode, &curKey)
		storage.calcCreateHeight(account, allocAccountStorageKeys, supernode, &curKey)
		storage.calcUpdateHeight(account, allocAccountStorageKeys, supernode, &curKey)
	}
}

func (storage *SuperNodeStorageStorage) calcId(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_address(102, supernode.Addr))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, supernode.Id)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *SuperNodeStorageStorage) calcName(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, supernode.Name)
	if len(storageKeys) != len(storageValues) {
		panic("get storage failed")
	}
	for i := range storageKeys {
		account.Storage[storageKeys[i]] = storageValues[i]
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKeys[i])
	}
}

func (storage *SuperNodeStorageStorage) calcAddr(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Addr(*curKey, supernode.Addr)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *SuperNodeStorageStorage) calcCreator(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Addr(*curKey, supernode.Creator)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *SuperNodeStorageStorage) calcEnode(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, supernode.Enode)
	if len(storageKeys) != len(storageValues) {
		panic("get storage failed")
	}
	for i := range storageKeys {
		account.Storage[storageKeys[i]] = storageValues[i]
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKeys[i])
	}
}

func (storage *SuperNodeStorageStorage) calcDesc(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, supernode.Description)
	if len(storageKeys) != len(storageValues) {
		panic("get storage failed")
	}
	for i := range storageKeys {
		account.Storage[storageKeys[i]] = storageValues[i]
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKeys[i])
	}
}

func (storage *SuperNodeStorageStorage) calcIsOfficial(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Bool(*curKey, supernode.IsOfficial)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *SuperNodeStorageStorage) calcStateInfo(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	var storageKey, storageValue common.Hash
	// state
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue = utils.GetStorage4Int(*curKey, supernode.StateInfo.State)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	// height
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue = utils.GetStorage4Int(*curKey, supernode.StateInfo.Height)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *SuperNodeStorageStorage) calcFounders(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey := common.BigToHash(*curKey)
	storageValue := common.BigToHash(big.NewInt(int64(len(supernode.Founders))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_bytes32(common.BigToHash(*curKey).Hex()))
	var subStorageKey, subStorageValue common.Hash
	for _, founder := range supernode.Founders {
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

func (storage *SuperNodeStorageStorage) calcIncentivePlan(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	var storageKey, storageValue common.Hash
	// creator
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue = utils.GetStorage4Int(*curKey, supernode.IncentivePlan.Creator)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	// partner
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue = utils.GetStorage4Int(*curKey, supernode.IncentivePlan.Partner)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	// voter
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue = utils.GetStorage4Int(*curKey, supernode.IncentivePlan.Voter)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *SuperNodeStorageStorage) calcVoteInfo(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	// voters
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey := common.BigToHash(*curKey)
	storageValue := common.BigToHash(big.NewInt(int64(len(supernode.VoteInfo.Voters))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_bytes32(common.BigToHash(*curKey).Hex()))
	var subStorageKey, subStorageValue common.Hash
	for _, voter := range supernode.VoteInfo.Voters {
		// lockID
		subStorageKey, subStorageValue = utils.GetStorage4Int(subKey, voter.LockID)
		account.Storage[subStorageKey] = subStorageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKey)
		subKey = subKey.Add(subKey, big.NewInt(1))

		// addr
		subStorageKey, subStorageValue = utils.GetStorage4Addr(subKey, voter.Addr)
		account.Storage[subStorageKey] = subStorageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKey)
		subKey = subKey.Add(subKey, big.NewInt(1))

		// amount
		subStorageKey, subStorageValue = utils.GetStorage4Int(subKey, voter.Amount)
		account.Storage[subStorageKey] = subStorageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKey)
		subKey = subKey.Add(subKey, big.NewInt(1))

		// height
		subStorageKey, subStorageValue = utils.GetStorage4Int(subKey, voter.Height)
		account.Storage[subStorageKey] = subStorageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKey)
		subKey = subKey.Add(subKey, big.NewInt(1))
	}

	// totalAmount
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue = utils.GetStorage4Int(*curKey, supernode.VoteInfo.TotalAmount)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	// totalNum
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue = utils.GetStorage4Int(*curKey, supernode.VoteInfo.TotalNum)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	// height
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue = utils.GetStorage4Int(*curKey, supernode.VoteInfo.Height)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *SuperNodeStorageStorage) calcLastRewardHeight(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, supernode.LastRewardHeight)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *SuperNodeStorageStorage) calcCreateHeight(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, supernode.CreateHeight)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *SuperNodeStorageStorage) calcUpdateHeight(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernode types.SuperNodeInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, supernode.UpdateHeight)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *SuperNodeStorageStorage) buildIDs(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernodes *[]types.SuperNodeInfo) {
	storageKey := common.BigToHash(big.NewInt(103))
	storageValue := common.BigToHash(big.NewInt(int64(len(*supernodes))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(103))
	for i, supernode := range *supernodes {
		curKey := big.NewInt(0).Add(subKey, big.NewInt(int64(i)))
		subStorageKey, subStorageValue := utils.GetStorage4Int(curKey, supernode.Id)
		account.Storage[subStorageKey] = subStorageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKey)
	}
}

func (storage *SuperNodeStorageStorage) buildID2Addr(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernodes *[]types.SuperNodeInfo) {
	for _, supernode := range *supernodes {
		curKey := big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(104, supernode.Id.Int64()))
		storageKey, storageValue := utils.GetStorage4Addr(curKey, supernode.Addr)
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
	}
}

func (storage *SuperNodeStorageStorage) buildName2Addr(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernodes *[]types.SuperNodeInfo) {
	for _, supernode := range *supernodes {
		curKey := big.NewInt(0).SetBytes(utils.Keccak256_uint_string(105, supernode.Name))
		storageKey, storageValue := utils.GetStorage4Addr(curKey, supernode.Addr)
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
	}
}

func (storage *SuperNodeStorageStorage) buildEnode2Addr(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, supernodes *[]types.SuperNodeInfo) {
	for _, supernode := range *supernodes {
		curKey := big.NewInt(0).SetBytes(utils.Keccak256_uint_string(106, supernode.Enode))
		storageKey, storageValue := utils.GetStorage4Addr(curKey, supernode.Addr)
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
	}
}

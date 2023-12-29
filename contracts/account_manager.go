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

type AccountManagerStorage struct {
	workPath  string
	ownerAddr string
}

func NewAccountManagerStorage(workPath string, ownerAddr string) *AccountManagerStorage {
	return &AccountManagerStorage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *AccountManagerStorage) Generate(alloc *core.GenesisAlloc) {
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

	contractNames := [2]string{"TransparentUpgradeableProxy", "AccountManager"}
	contractAddrs := [2]string{"0x0000000000000000000000000000000000001010", "0x0000000000000000000000000000000000001011"}

	for i := range contractNames {
		codePath := storage.workPath + "temp" + string(filepath.Separator) + contractNames[i] + ".bin-runtime"
		code, err := os.ReadFile(codePath)
		if err != nil {
			panic(err)
		}

		account := core.GenesisAccount{
			Balance: big.NewInt(0).String(),
			Code:    "0x" + string(code),
		}
		if contractNames[i] == "TransparentUpgradeableProxy" {
			account.Balance = totalAmount.String()
			account.Storage = make(map[common.Hash]common.Hash)
			account.Storage[common.BigToHash(big.NewInt(0))] = common.BigToHash(big.NewInt(1))
			account.Storage[common.BigToHash(big.NewInt(0x33))] = common.HexToHash(storage.ownerAddr)
			account.Storage[common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")] = common.HexToHash(contractAddrs[1])
			account.Storage[common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")] = common.HexToHash(ProxyAdminAddr)

			// record_no
			storage.buildRecordNo(&account, len(addrs))

			// addr2records
			storage.buildAddr2records(&account, addrs, amounts)

			// id2index
			storage.buildID2index(&account, addrs)

			// id2addr
			storage.buildID2addr(&account, addrs)

			// id2useinfo
			storage.buildID2useInfo(&account, addrs)
		}
		(*alloc)[common.HexToAddress(contractAddrs[i])] = account
	}
	os.RemoveAll(storage.workPath + "temp")
}

func (storage *AccountManagerStorage) LoadMasterNode() *[]types.MasterNodeInfo {
	jsonFile, err := os.Open(storage.workPath + utils.GetDataDir() + string(filepath.Separator) + "MasterNode.info")
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

func (storage *AccountManagerStorage) buildRecordNo(account *core.GenesisAccount, count int) {
	curKey := big.NewInt(102)
	storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(int64(count)))
	account.Storage[storageKey] = storageValue
}

func (storage *AccountManagerStorage) buildAddr2records(account *core.GenesisAccount, addrs []common.Address, amounts []*big.Int) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for i, addr := range addrs {
		// size
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_address(103, addr))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(1))
		account.Storage[storageKey] = storageValue

		// id
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_bytes32(common.BigToHash(curKey).Hex()))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(int64(i+1)))
		account.Storage[storageKey] = storageValue
		// addr
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Addr(curKey, addr)
		account.Storage[storageKey] = storageValue
		// amount
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, amounts[i])
		account.Storage[storageKey] = storageValue
		// lockDay
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(720))
		account.Storage[storageKey] = storageValue
		// startHeight
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(0))
		account.Storage[storageKey] = storageValue
		// unlockHeight
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(720*24*3600/30))
		account.Storage[storageKey] = storageValue
	}
}

func (storage *AccountManagerStorage) buildID2index(account *core.GenesisAccount, addrs []common.Address) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for i := range addrs {
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(104, int64(i+1)))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(0))
		account.Storage[storageKey] = storageValue
	}
}

func (storage *AccountManagerStorage) buildID2addr(account *core.GenesisAccount, addrs []common.Address) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for i, addr := range addrs {
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(105, int64(i+1)))
		storageKey, storageValue = utils.GetStorage4Addr(curKey, addr)
		account.Storage[storageKey] = storageValue
	}
}

func (storage *AccountManagerStorage) buildID2useInfo(account *core.GenesisAccount, addrs []common.Address) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for i, addr := range addrs {
		// frozenAddr
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(106, int64(i+1)))
		storageKey, storageValue = utils.GetStorage4Addr(curKey, addr)
		account.Storage[storageKey] = storageValue
		// freezeHeight
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(0))
		account.Storage[storageKey] = storageValue
		// unfreezeHeight
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(720*24*3600/30))
		account.Storage[storageKey] = storageValue
		// votedAddr
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Addr(curKey, common.Address{})
		account.Storage[storageKey] = storageValue
		// voteHeight
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(0))
		account.Storage[storageKey] = storageValue
		// releaseHeight
		curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(0))
		account.Storage[storageKey] = storageValue
	}
}

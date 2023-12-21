package contracts

import (
	"bufio"
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
	"strings"
)

type Safe3Storage struct {
	workPath  string
	ownerAddr common.Address
}

func NewSafe3Storage(workPath string, ownerAddr common.Address) *Safe3Storage {
	return &Safe3Storage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *Safe3Storage) Generate(genesis *core.Genesis, allocAccounts *[]common.Address, mapAllocAccountStorageKeys *map[common.Address][]common.Hash) {
	utils.Compile(storage.workPath, "Safe3.sol")

	lockInfos, lockAmounts, lockNum := storage.loadLockedInfos()

	totalAmount := big.NewInt(0)
	infos := storage.loadBalance(lockInfos, lockAmounts, totalAmount)

	contractNames := [2]string{"TransparentUpgradeableProxy", "Safe3"}
	contractAddrs := [2]string{"0x0000000000000000000000000000000000001090", "0x0000000000000000000000000000000000001091"}

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
			account.Balance = totalAmount
			account.Storage = make(map[common.Hash]common.Hash)

			account.Storage[common.BigToHash(big.NewInt(0))] = common.BigToHash(big.NewInt(1))
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.BigToHash(big.NewInt(0)))

			account.Storage[common.BigToHash(big.NewInt(0x33))] = common.HexToHash(storage.ownerAddr.Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.BigToHash(big.NewInt(0x33)))

			account.Storage[common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")] = common.HexToHash(common.HexToAddress(contractAddrs[1]).Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc"))

			account.Storage[common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")] = common.HexToHash(ProxyAdminAddr.Hex())
			allocAccountStorageKeys = append(allocAccountStorageKeys, common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103"))

			// num
			storage.buildNum(&account, &allocAccountStorageKeys, infos)

			// keyIDs
			storage.buildKeyIDs(&account, &allocAccountStorageKeys, infos)

			// availables
			storage.buildAvailables(&account, &allocAccountStorageKeys, infos)

			// lockedNum
			storage.buildLockedNum(&account, &allocAccountStorageKeys, lockNum)

			// lockedKeyIDs
			storage.buildLockedKeyIDs(&account, &allocAccountStorageKeys, lockInfos)

			// locks
			storage.buildLocks(&account, &allocAccountStorageKeys, lockInfos)
		}

		if len(allocAccountStorageKeys) != 0 {
			(*mapAllocAccountStorageKeys)[addr] = allocAccountStorageKeys
		}

		genesis.Alloc[addr] = account
	}
	os.RemoveAll(storage.workPath + "temp")
}

func (storage *Safe3Storage) loadBalance(lockInfos map[string][]types.Safe3LockInfo, lockAmounts map[string]*big.Int, totalAmount *big.Int) *[]types.Safe3Info {
	file, err := os.Open(storage.workPath + utils.GetDataDir() + string(filepath.Separator) + "safe3" + string(filepath.Separator) + "balanceaddress.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	CHANGE_COIN := big.NewInt(10000000000)
	MIN_COIN := big.NewInt(100000000000000000) // 0.1 safe

	infos := new([]types.Safe3Info)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, `"`, ``, -1)
		temps := strings.Split(line, ",")
		if len(temps) != 6 || len(temps[1]) != 34 {
			continue
		}
		addr := temps[1]
		amount, _ := new(big.Int).SetString(temps[2], 10)
		amount.Mul(amount, CHANGE_COIN)
		if amount.Cmp(MIN_COIN) < 0 {
			continue
		}
		totalAmount.Add(totalAmount, amount)
		lockAmount, _ := new(big.Int).SetString(temps[3], 10)
		lockAmount.Mul(lockAmount, CHANGE_COIN)
		if lockAmounts[addr] != nil && lockAmount.Cmp(lockAmounts[addr]) <= 0 {
			amount.Sub(amount, lockAmounts[addr])
		}
		*infos = append(*infos, types.Safe3Info{
			Addr:         addr,
			Amount:       amount,
			RedeemHeight: big.NewInt(0)})
	}
	return infos
}

func (storage *Safe3Storage) loadMNs() map[string]string {
	jsonFile, err := os.Open(storage.workPath + utils.GetDataDir() + string(filepath.Separator) + "safe3" + string(filepath.Separator) + "masternodes.info")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	masternodes := new(map[string]string)
	err = json.Unmarshal(jsonData, masternodes)
	if err != nil {
		panic(err)
	}
	return *masternodes
}

func (storage *Safe3Storage) loadLockedInfos() (map[string][]types.Safe3LockInfo, map[string]*big.Int, int64) {
	masternodes := storage.loadMNs()

	file, err := os.Open(storage.workPath + utils.GetDataDir() + string(filepath.Separator) + "safe3" + string(filepath.Separator) + "lockedaddresses.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	lockNum := int64(0)
	lockInfos := make(map[string][]types.Safe3LockInfo)
	lockAmounts := make(map[string]*big.Int)

	ETH_COIN := new(big.Float).SetInt(big.NewInt(1000000000000000000))
	MIN_COIN := big.NewInt(100000000000000000) // 0.1 safe
	SAFE3_END_HEIGHT := big.NewInt(5000000)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, `"`, ``, -1)
		temps := strings.Split(line, ",")
		if len(temps) != 7 || len(temps[2]) != 34 {
			continue
		}
		txid := temps[0][35:]
		txid = strings.Replace(txid, "_", "-", 1)
		addr := temps[2]
		amt, _ := new(big.Float).SetString(temps[4])
		amt.Mul(amt, ETH_COIN)
		amount := new(big.Int)
		amt.Int(amount)
		if amount.Cmp(MIN_COIN) < 0 {
			continue
		}
		lockHeight, _ := new(big.Int).SetString(temps[5], 10)
		unlockHeight, _ := new(big.Int).SetString(temps[6], 10)
		isMN := false
		mnState := big.NewInt(0)
		if len(masternodes[txid]) != 0 {
			isMN = true
			if masternodes[txid] == "ENABLED" {
				mnState = big.NewInt(1)
			} else {
				mnState = big.NewInt(2)
			}
		} else {
			if unlockHeight.Cmp(SAFE3_END_HEIGHT) <= 0 {
				continue
			}
		}

		lockNum++
		if lockAmounts[addr] == nil {
			lockAmounts[addr] = amount
		} else {
			lockAmounts[addr].Add(lockAmounts[addr], amount)
		}
		lockInfos[addr] = append(lockInfos[addr], types.Safe3LockInfo{
			Addr:         addr,
			Amount:       amount,
			LockHeight:   lockHeight,
			UnlockHeight: unlockHeight,
			Txid:         txid,
			IsMN:         isMN,
			MnState:      mnState,
			RedeemHeight: big.NewInt(0)})
	}
	return lockInfos, lockAmounts, lockNum
}

func (storage *Safe3Storage) buildNum(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, infos *[]types.Safe3Info) {
	curKey := big.NewInt(101)
	storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(int64(len(*infos))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *Safe3Storage) buildKeyIDs(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, infos *[]types.Safe3Info) {
	storageKey := common.BigToHash(big.NewInt(102))
	storageValue := common.BigToHash(big.NewInt(int64(len(*infos))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(102))
	for i, info := range *infos {
		curKey := big.NewInt(0).Add(subKey, big.NewInt(int64(i)))
		keyID := getKeyIDFromAddress(info.Addr)
		subStorageKeys, subStorageValues := utils.GetStorage4Bytes(curKey, keyID)
		if len(subStorageKeys) != len(subStorageValues) {
			panic("get storage failed")
		}
		for k := range subStorageKeys {
			account.Storage[subStorageKeys[k]] = subStorageValues[k]
			*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKeys[k])
		}
	}
}

func (storage *Safe3Storage) buildAvailables(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, infos *[]types.Safe3Info) {
	var curKey *big.Int
	for _, info := range *infos {
		storage.calcAddr(account, allocAccountStorageKeys, info, &curKey)
		storage.calcAmount(account, allocAccountStorageKeys, info, &curKey)
		storage.calcRedeemHeight(account, allocAccountStorageKeys, info, &curKey)
	}
}

func (storage *Safe3Storage) calcAddr(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, info types.Safe3Info, curKey **big.Int) {
	*curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_bytes(103, getKeyIDFromAddress(info.Addr)))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, info.Addr)
	if len(storageKeys) != len(storageValues) {
		panic("get storage failed")
	}
	for i := range storageKeys {
		account.Storage[storageKeys[i]] = storageValues[i]
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKeys[i])
	}
}

func (storage *Safe3Storage) calcAmount(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, info types.Safe3Info, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, info.Amount)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *Safe3Storage) calcRedeemHeight(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, info types.Safe3Info, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, info.RedeemHeight)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *Safe3Storage) buildLockedNum(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, lockNum int64) {
	curKey := big.NewInt(104)
	storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(lockNum))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *Safe3Storage) buildLockedKeyIDs(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, infos map[string][]types.Safe3LockInfo) {
	storageKey := common.BigToHash(big.NewInt(105))
	storageValue := common.BigToHash(big.NewInt(int64(len(infos))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(105))

	i := int64(0)
	for addr, _ := range infos {
		curKey := big.NewInt(0).Add(subKey, big.NewInt(i))
		keyID := getKeyIDFromAddress(addr)
		subStorageKeys, subStorageValues := utils.GetStorage4Bytes(curKey, keyID)
		if len(subStorageKeys) != len(subStorageValues) {
			panic("get storage failed")
		}
		for k := range subStorageKeys {
			account.Storage[subStorageKeys[k]] = subStorageValues[k]
			*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKeys[k])
		}
		i++
	}
}

func (storage *Safe3Storage) buildLocks(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, infos map[string][]types.Safe3LockInfo) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for addr, list := range infos {
		// size
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_bytes(106, getKeyIDFromAddress(addr)))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(int64(len(list))))
		account.Storage[storageKey] = storageValue
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

		curKey = big.NewInt(0).SetBytes(utils.Keccak256_bytes32(common.BigToHash(curKey).Hex()))
		curKey = curKey.Sub(curKey, big.NewInt(1))
		for _, info := range list {
			storage.calcAddr2(account, allocAccountStorageKeys, info, &curKey)
			storage.calcAmount2(account, allocAccountStorageKeys, info, &curKey)
			storage.calcLockHeight(account, allocAccountStorageKeys, info, &curKey)
			storage.calcUnlockHeight(account, allocAccountStorageKeys, info, &curKey)
			storage.calcTxid(account, allocAccountStorageKeys, info, &curKey)
			storage.calcIsMN(account, allocAccountStorageKeys, info, &curKey)
			storage.calcMNState(account, allocAccountStorageKeys, info, &curKey)
			storage.calcRedeemHeight2(account, allocAccountStorageKeys, info, &curKey)
		}
	}
}

func (storage *Safe3Storage) calcAddr2(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, info types.Safe3LockInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, info.Addr)
	if len(storageKeys) != len(storageValues) {
		panic("get storage failed")
	}
	for i := range storageKeys {
		account.Storage[storageKeys[i]] = storageValues[i]
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKeys[i])
	}
}

func (storage *Safe3Storage) calcAmount2(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, info types.Safe3LockInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, info.Amount)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *Safe3Storage) calcLockHeight(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, info types.Safe3LockInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, info.LockHeight)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *Safe3Storage) calcUnlockHeight(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, info types.Safe3LockInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, info.UnlockHeight)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *Safe3Storage) calcTxid(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, info types.Safe3LockInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, info.Txid)
	if len(storageKeys) != len(storageValues) {
		panic("get storage failed")
	}
	for i := range storageKeys {
		account.Storage[storageKeys[i]] = storageValues[i]
		*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKeys[i])
	}
}

func (storage *Safe3Storage) calcIsMN(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, info types.Safe3LockInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Bool(*curKey, info.IsMN)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *Safe3Storage) calcMNState(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, info types.Safe3LockInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, info.MnState)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *Safe3Storage) calcRedeemHeight2(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, info types.Safe3LockInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, info.RedeemHeight)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func getKeyIDFromAddress(addr string) []byte {
	return utils.Base58Decoding(addr)
}

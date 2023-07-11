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
	workPath string
	ownerAddr common.Address
}

func NewSafe3Storage(workPath string, ownerAddr common.Address) *Safe3Storage {
	return &Safe3Storage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *Safe3Storage) Generate(genesis *core.Genesis, allocAccounts *[]common.Address, mapAllocAccountStorageKeys *map[common.Address][]common.Hash) {
	utils.Compile(storage.workPath, "Safe3.sol")

	//infos := storage.loadInfos()
	//lockInfos := storage.loadLockedInfos()

	contractNames := [3]string{"Safe3", "ProxyAdmin", "TransparentUpgradeableProxy"}
	contractAddrs := [3]string{"0x0000000000000000000000000000000000001090", "0x0000000000000000000000000000000000001091", "0x0000000000000000000000000000000000001092"}

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

			//// num
			//storage.buildNum(&account, &allocAccountStorageKeys, infos)
			//
			//// addrs
			//storage.buildKeyIDs(&account, &allocAccountStorageKeys, infos)
			//
			//// availables
			//storage.buildAvailables(&account, &allocAccountStorageKeys, infos)
			//
			//// lockNum
			//storage.buildLockedNum(&account, &allocAccountStorageKeys, lockInfos)
			//
			//// lockedAddrs
			//storage.buildLockedKeyIDs(&account, &allocAccountStorageKeys, lockInfos)
			//
			//// locks
			//storage.buildLocks(&account, &allocAccountStorageKeys, lockInfos)
		}

		if len(allocAccountStorageKeys) != 0 {
			(*mapAllocAccountStorageKeys)[addr] = allocAccountStorageKeys
		}

		genesis.Alloc[addr] = account
	}
	os.RemoveAll(storage.workPath + "temp")
}

func (storage *Safe3Storage) loadInfos() *[]types.Safe3Info {
	file, err := os.Open(storage.workPath + "data" + string(filepath.Separator) + "safe3" + string(filepath.Separator) + "availables.info")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	infos := new([]types.Safe3Info)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		temps := strings.Split(line, "\t")
		if len(temps) != 2 {
			continue
		}
		addr := strings.TrimSpace(temps[0])
		amount, _ := new(big.Int).SetString(strings.TrimSpace(temps[1]), 10)
		amount.Mul(amount, big.NewInt(10000000000))
		if amount.Uint64() < 100000000000000000 {
			continue
		}
		*infos = append(*infos, types.Safe3Info{Addr: addr,
												Amount: amount,
												RedeemHeight: big.NewInt(0)})
	}
	return infos
}

func (storage *Safe3Storage) loadMNs() map[string]string {
	jsonFile, err := os.Open(storage.workPath + "data" + string(filepath.Separator) + "safe3" + string(filepath.Separator) + "masternodes.info")
	if err != nil {
		panic(err)
	}
	defer jsonFile.Close()

	jsonData, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		panic(err)
	}

	temps := new(map[string]string)
	err = json.Unmarshal(jsonData, temps)
	if err != nil {
		panic(err)
	}

	masternodes := make(map[string]string)
	for key, value := range *temps {
		txid := strings.Split(key, "-")[0]
		if len(masternodes[txid]) != 0 {
			continue
		}
		masternodes[txid] = value
	}
	return masternodes
}

func (storage *Safe3Storage) loadLockedInfos() *[]types.Safe3LockInfo {
	masternodes := storage.loadMNs()

	file, err := os.Open(storage.workPath + "data" + string(filepath.Separator) + "safe3" + string(filepath.Separator) + "locks.info")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	infos := new([]types.Safe3LockInfo)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		temps := strings.Split(line, "\t")
		if len(temps) != 5 {
			continue
		}
		addr := strings.TrimSpace(temps[0])
		txid := strings.TrimSpace(temps[1])
		amt, _ := new(big.Float).SetString(strings.TrimSpace(temps[2]))
		coin := new(big.Float).SetInt(big.NewInt(1000000000000000000))
		amt.Mul(amt, coin)
		amount := new(big.Int)
		amt.Int(amount)
		if amount.Uint64() < 100000000000000000 {
			continue
		}
		lockHeight, _ := new(big.Int).SetString(strings.TrimSpace(temps[3]), 10)
		unlockHeight, _ := new(big.Int).SetString(strings.TrimSpace(temps[4]), 10)
		isMN := false
		if amount.Cmp(big.NewInt(1000000000)) >= 0 && len(masternodes[txid]) != 0 {
			isMN = true
		}
		*infos = append(*infos, types.Safe3LockInfo{Addr: addr,
													Amount: amount,
													LockHeight: lockHeight,
													UnlockHeight: unlockHeight,
													Txid: txid,
													IsMN: isMN,
													RedeemHeight: big.NewInt(0)})
	}
	return infos
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
		for k, _ := range subStorageKeys {
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
	*curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_string(103, info.Addr))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, info.Addr)
	if len(storageKeys) != len(storageValues) {
		panic("get storage failed")
	}
	for i, _ := range storageKeys {
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

func (storage *Safe3Storage) buildLockedNum(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, infos *[]types.Safe3LockInfo) {
	curKey := big.NewInt(104)
	storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(int64(len(*infos))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func (storage *Safe3Storage) buildLockedKeyIDs(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, infos *[]types.Safe3LockInfo) {
	storageKey := common.BigToHash(big.NewInt(102))
	storageValue := common.BigToHash(big.NewInt(int64(len(*infos))))
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(105))
	for i, info := range *infos {
		curKey := big.NewInt(0).Add(subKey, big.NewInt(int64(i)))
		keyID := getKeyIDFromAddress(info.Addr)
		subStorageKeys, subStorageValues := utils.GetStorage4Bytes(curKey, keyID)
		if len(subStorageKeys) != len(subStorageValues) {
			panic("get storage failed")
		}
		for k, _ := range subStorageKeys {
			account.Storage[subStorageKeys[k]] = subStorageValues[k]
			*allocAccountStorageKeys = append(*allocAccountStorageKeys, subStorageKeys[k])
		}
	}
}

func (storage *Safe3Storage) buildLocks(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, infos *[]types.Safe3LockInfo) {
	var curKey *big.Int
	for _, info := range *infos {
		storage.calcAddr2(account, allocAccountStorageKeys, info, &curKey)
		storage.calcAmount2(account, allocAccountStorageKeys, info, &curKey)
		storage.calcLockHeight(account, allocAccountStorageKeys, info, &curKey)
		storage.calcUnlockHeight(account, allocAccountStorageKeys, info, &curKey)
		storage.calcTxid(account, allocAccountStorageKeys, info, &curKey)
		storage.calcIsMN(account, allocAccountStorageKeys, info, &curKey)
		storage.calcRedeemHeight2(account, allocAccountStorageKeys, info, &curKey)
	}
}

func (storage *Safe3Storage) calcAddr2(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, info types.Safe3LockInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_string(106, info.Addr))
	storageKeys, storageValues := utils.GetStorage4String(*curKey, info.Addr)
	if len(storageKeys) != len(storageValues) {
		panic("get storage failed")
	}
	for i, _ := range storageKeys {
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
	for i, _ := range storageKeys {
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

func (storage *Safe3Storage) calcRedeemHeight2(account *core.GenesisAccount, allocAccountStorageKeys *[]common.Hash, info types.Safe3LockInfo, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Int(*curKey, info.RedeemHeight)
	account.Storage[storageKey] = storageValue
	*allocAccountStorageKeys = append(*allocAccountStorageKeys, storageKey)
}

func getKeyIDFromAddress(addr string) []byte {
	b := utils.Base58Decoding(addr)
	return b[1:21]
}
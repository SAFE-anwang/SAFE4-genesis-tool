package contracts

import (
	"bufio"
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
	ownerAddr string
}

func NewSafe3Storage(workPath string, ownerAddr string) *Safe3Storage {
	return &Safe3Storage{workPath: workPath, ownerAddr: ownerAddr}
}

func (storage *Safe3Storage) Generate(alloc *core.GenesisAlloc) {
	utils.Compile(storage.workPath, "Safe3.sol")

	lockedInfos, lockedAmounts, lockedNum := storage.loadLockedInfos()

	totalAmount := big.NewInt(0)
	specialAmounts := storage.loadSpecialInfos()
	availableAmounts := storage.loadBalance(lockedAmounts, specialAmounts, totalAmount)

	contractNames := [2]string{"TransparentUpgradeableProxy", "Safe3"}
	contractAddrs := [2]string{"0x0000000000000000000000000000000000001090", "0x0000000000000000000000000000000000001091"}

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

			// keyIDs
			storage.buildKeyIDs(&account, availableAmounts)

			// availables
			storage.buildAvailables(&account, availableAmounts)

			// lockedNum
			storage.buildLockedNum(&account, lockedNum)

			// lockedKeyIDs
			storage.buildLockedKeyIDs(&account, lockedInfos)

			// locks
			storage.buildLocks(&account, lockedInfos)

			// specialKeyIDs
			storage.buildSpecialKeyIDs(&account, specialAmounts)

			// availables
			storage.buildSpecials(&account, specialAmounts)
		}
		(*alloc)[common.HexToAddress(contractAddrs[i])] = account
	}
	os.RemoveAll(storage.workPath + "temp")
}

func (storage *Safe3Storage) loadBalance(lockedAmounts map[string]*big.Int, specialAmounts map[string]*big.Int, totalAmount *big.Int) map[string]*big.Int {
	file, err := os.Open(storage.workPath + utils.GetDataDir() + string(filepath.Separator) + "safe3" + string(filepath.Separator) + "balanceaddress.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	CHANGE_COIN := big.NewInt(10000000000)
	MIN_COIN := big.NewInt(100000000000000000 - 1) // 0.1 safe

	availableAmounts := make(map[string]*big.Int)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, `"`, ``, -1)
		temps := strings.Split(line, ",")
		if len(temps) < 6 || len(temps[1]) != 34 {
			continue
		}
		addr := temps[1]

		amount, _ := new(big.Int).SetString(temps[2], 10)
		amount.Mul(amount, CHANGE_COIN)
		if amount.Cmp(MIN_COIN) < 0 {
			continue
		}
		totalAmount.Add(totalAmount, amount)

		if specialAmounts[addr] != nil {
			continue
		}

		lockedAmount, _ := new(big.Int).SetString(temps[3], 10)
		lockedAmount.Mul(lockedAmount, CHANGE_COIN)
		if lockedAmounts[addr] != nil && lockedAmount.Cmp(lockedAmounts[addr]) <= 0 {
			lockedAmount = lockedAmounts[addr]
		}
		availableAmounts[addr] = big.NewInt(0).Sub(amount, lockedAmount)
	}
	return availableAmounts
}

func (storage *Safe3Storage) loadSpecialInfos() map[string]*big.Int {
	file, err := os.Open(storage.workPath + utils.GetDataDir() + string(filepath.Separator) + "safe3" + string(filepath.Separator) + "specialaddress.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	CHANGE_COIN := big.NewInt(10000000000)
	MIN_COIN := big.NewInt(100000000000000000 - 1) // 0.1 safe

	specialAmounts := make(map[string]*big.Int)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, `"`, ``, -1)
		temps := strings.Split(line, ",")
		if len(temps) < 6 || len(temps[1]) != 34 {
			continue
		}
		addr := temps[1]
		amount, _ := new(big.Int).SetString(temps[2], 10)
		amount.Mul(amount, CHANGE_COIN)
		if amount.Cmp(MIN_COIN) < 0 {
			continue
		}
		specialAmounts[addr] = amount
	}
	return specialAmounts
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

func (storage *Safe3Storage) loadLockedInfos() (map[string][]types.LockedData, map[string]*big.Int, int64) {
	masternodes := storage.loadMNs()

	file, err := os.Open(storage.workPath + utils.GetDataDir() + string(filepath.Separator) + "safe3" + string(filepath.Separator) + "lockedaddresses.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	lockedInfos := make(map[string][]types.LockedData)
	lockedAmounts := make(map[string]*big.Int)
	lockedNum := int64(0)

	ETH_COIN := new(big.Float).SetInt(big.NewInt(1000000000000000000))
	MIN_COIN := big.NewInt(100000000000000000 - 1) // 0.1 safe
	SAFE3_END_HEIGHT := big.NewInt(5000000)
	SPOS_HEIGHT := big.NewInt(1092826)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, `"`, ``, -1)
		temps := strings.Split(line, ",")
		if len(temps) < 7 || len(temps[2]) != 34 {
			continue
		}
		txid := temps[0][35:99]
		n, _ := big.NewInt(0).SetString(temps[0][100:], 10)
		addr := temps[2]
		amt, _ := new(big.Float).SetString(temps[4])
		amt.Mul(amt, ETH_COIN)
		amount, _ := amt.Int(nil)
		lockHeight, _ := new(big.Int).SetString(temps[5], 10)
		unlockHeight, _ := new(big.Int).SetString(temps[6], 10)
		lockDay := big.NewInt(0)
		remainLockHeight := big.NewInt(0)
		isMN := false
		if len(masternodes[txid+"-"+temps[0][100:]]) != 0 { // masternode
			isMN = true
			lockDay = big.NewInt(90)              // add 3 months
			remainLockHeight = big.NewInt(259200) // 3 months
		} else {
			if unlockHeight.Cmp(SAFE3_END_HEIGHT) < 0 { // unlocked common-lock
				continue
			}
			if amount.Cmp(MIN_COIN) < 0 {
				continue
			}
		}
		if unlockHeight.Cmp(SAFE3_END_HEIGHT) >= 0 {
			day := int64(0)
			if lockHeight.Cmp(SPOS_HEIGHT) < 0 {
				day += (unlockHeight.Int64() - lockHeight.Int64()) / 576
				if (unlockHeight.Int64()-lockHeight.Int64())%576 != 0 {
					day += 1
				}
			} else {
				day += (unlockHeight.Int64() - lockHeight.Int64()) / 2880
				if (unlockHeight.Int64()-lockHeight.Int64())%2880 != 0 {
					day += 1
				}
			}
			lockDay.Add(lockDay, big.NewInt(day))
			remainLockHeight.Add(remainLockHeight, big.NewInt(unlockHeight.Int64()-SAFE3_END_HEIGHT.Int64()))
		}

		if isMN && lockDay.Int64() < 720 {
			lockDay = big.NewInt(720)
		}

		lockedNum++
		if lockedAmounts[addr] == nil {
			lockedAmounts[addr] = amount
		} else {
			lockedAmounts[addr] = big.NewInt(0).Add(lockedAmounts[addr], amount)
		}
		lockedInfos[addr] = append(lockedInfos[addr], types.LockedData{
			Txid:             common.HexToHash(txid),
			N:                n,
			Amount:           amount,
			LockHeight:       lockHeight,
			UnlockHeight:     unlockHeight,
			RemainLockHeight: remainLockHeight,
			LockDay:          lockDay,
			IsMN:             isMN,
		})
	}
	return lockedInfos, lockedAmounts, lockedNum
}

func (storage *Safe3Storage) buildKeyIDs(account *core.GenesisAccount, availableAmounts map[string]*big.Int) {
	storageKey := common.BigToHash(big.NewInt(101))
	storageValue := common.BigToHash(big.NewInt(int64(len(availableAmounts))))
	account.Storage[storageKey] = storageValue

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(101))
	index := int64(0)
	for addr := range availableAmounts {
		curKey := big.NewInt(0).Add(subKey, big.NewInt(index))
		keyID := getKeyIDFromAddress(addr)
		subStorageKeys, subStorageValues := utils.GetStorage4Bytes(curKey, keyID)
		for k := range subStorageKeys {
			account.Storage[subStorageKeys[k]] = subStorageValues[k]
		}
		index++
	}
}

func (storage *Safe3Storage) buildAvailables(account *core.GenesisAccount, availableAmounts map[string]*big.Int) {
	var curKey *big.Int
	for addr, amount := range availableAmounts {
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_bytes(102, getKeyIDFromAddress(addr)))
		storage.calcAmount(account, amount, &curKey)
	}
}

func (storage *Safe3Storage) calcAmount(account *core.GenesisAccount, amount *big.Int, curKey **big.Int) {
	storageKey, storageValue := utils.GetStorage4Int(*curKey, amount)
	account.Storage[storageKey] = storageValue
}

func (storage *Safe3Storage) buildLockedNum(account *core.GenesisAccount, lockedNum int64) {
	curKey := big.NewInt(103)
	storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(lockedNum))
	account.Storage[storageKey] = storageValue
}

func (storage *Safe3Storage) buildLockedKeyIDs(account *core.GenesisAccount, infos map[string][]types.LockedData) {
	storageKey := common.BigToHash(big.NewInt(104))
	storageValue := common.BigToHash(big.NewInt(int64(len(infos))))
	account.Storage[storageKey] = storageValue

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(104))

	i := int64(0)
	for addr := range infos {
		curKey := big.NewInt(0).Add(subKey, big.NewInt(i))
		keyID := getKeyIDFromAddress(addr)
		subStorageKeys, subStorageValues := utils.GetStorage4Bytes(curKey, keyID)
		for k := range subStorageKeys {
			account.Storage[subStorageKeys[k]] = subStorageValues[k]
		}
		i++
	}
}

func (storage *Safe3Storage) buildLocks(account *core.GenesisAccount, infos map[string][]types.LockedData) {
	var curKey *big.Int
	var storageKey, storageValue common.Hash

	for addr, list := range infos {
		// size
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_bytes(105, getKeyIDFromAddress(addr)))
		storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(int64(len(list))))
		account.Storage[storageKey] = storageValue

		curKey = big.NewInt(0).SetBytes(utils.Keccak256_bytes32(common.BigToHash(curKey).Hex()))
		curKey = curKey.Sub(curKey, big.NewInt(1))
		for _, info := range list {
			storage.calcTxid(account, info, &curKey)
			storage.calcPart(account, info, &curKey)
			curKey = curKey.Add(curKey, big.NewInt(1))
		}
	}
}

func (storage *Safe3Storage) calcTxid(account *core.GenesisAccount, info types.LockedData, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey, storageValue := utils.GetStorage4Bytes32(*curKey, info.Txid)
	account.Storage[storageKey] = storageValue
}

func (storage *Safe3Storage) calcPart(account *core.GenesisAccount, info types.LockedData, curKey **big.Int) {
	*curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
	storageKey := common.BigToHash(*curKey)

	nStorageValue := common.BigToHash(info.N) // n: 2 bytes

	amountStorageValue := common.BigToHash(info.Amount) // amount: 12 bytes

	lockHeightStorageValue := common.BigToHash(info.LockHeight) // lockHeight: 3 bytes

	unlockHeightStorageValue := common.BigToHash(info.UnlockHeight) // unlockHeight: 3 bytes

	remainLockHeightStorageValue := common.BigToHash(info.RemainLockHeight) // remainLockHeight: 3 bytes

	lockDayStorageValue := common.BigToHash(info.LockDay) // lockDay: 2 bytes

	isMNStorageValue := common.Hash{} // isMN: 1 bytes
	if info.IsMN {
		isMNStorageValue = common.BigToHash(big.NewInt(1))
	} else {
		isMNStorageValue = common.BigToHash(big.NewInt(0))
	}

	storageValue := common.Hash{}
	offset := 0
	for i := 0; i < 2; i++ {
		storageValue[31-i-offset] = nStorageValue[31-i]
	}
	offset += 2
	for i := 0; i < 12; i++ {
		storageValue[31-i-offset] = amountStorageValue[31-i]
	}
	offset += 12
	for i := 0; i < 3; i++ {
		storageValue[31-i-offset] = lockHeightStorageValue[31-i]
	}
	offset += 3
	for i := 0; i < 3; i++ {
		storageValue[31-i-offset] = unlockHeightStorageValue[31-i]
	}
	offset += 3
	for i := 0; i < 3; i++ {
		storageValue[31-i-offset] = remainLockHeightStorageValue[31-i]
	}
	offset += 3
	for i := 0; i < 2; i++ {
		storageValue[31-i-offset] = lockDayStorageValue[31-i]
	}
	offset += 2
	storageValue[31-offset] = isMNStorageValue[31]

	account.Storage[storageKey] = storageValue
}

func (storage *Safe3Storage) buildSpecialKeyIDs(account *core.GenesisAccount, specialAmounts map[string]*big.Int) {
	storageKey := common.BigToHash(big.NewInt(106))
	storageValue := common.BigToHash(big.NewInt(int64(len(specialAmounts))))
	account.Storage[storageKey] = storageValue

	subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(106))
	index := int64(0)
	for addr := range specialAmounts {
		curKey := big.NewInt(0).Add(subKey, big.NewInt(index))
		keyID := getKeyIDFromAddress(addr)
		subStorageKeys, subStorageValues := utils.GetStorage4Bytes(curKey, keyID)
		for k := range subStorageKeys {
			account.Storage[subStorageKeys[k]] = subStorageValues[k]
		}
		index++
	}
}

func (storage *Safe3Storage) buildSpecials(account *core.GenesisAccount, specialAmounts map[string]*big.Int) {
	var curKey *big.Int
	for addr, amount := range specialAmounts {
		curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_bytes(107, getKeyIDFromAddress(addr)))
		storage.calcAmount3(account, amount, &curKey)
	}
}

func (storage *Safe3Storage) calcAmount3(account *core.GenesisAccount, amount *big.Int, curKey **big.Int) {
	storageKey, storageValue := utils.GetStorage4Int(*curKey, amount)
	account.Storage[storageKey] = storageValue
}

func getKeyIDFromAddress(addr string) []byte {
	return utils.Base58Decoding(addr)
}

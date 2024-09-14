package contracts

import (
    "archive/zip"
    "bufio"
    "bytes"
    "compress/gzip"
    "encoding/json"
    "fmt"
    "github.com/safe/SAFE4-genesis-tool/common"
    "github.com/safe/SAFE4-genesis-tool/types"
    "github.com/safe/SAFE4-genesis-tool/utils"
    "io"
    "io/ioutil"
    "math/big"
    "os"
    "path/filepath"
    "strconv"
    "strings"
)

var fileIndex = 0
var storageList []string
var maxNum = 10240
var pairs = make(map[common.Hash]common.Hash)

//var MIN_COIN = big.NewInt(99999999) // 1 safe
var MIN_COIN = big.NewInt(9999999) // 0.1 safe
//var MIN_COIN = big.NewInt(999999) // 0.01 safe
//var MIN_COIN = big.NewInt(99999) // 0.001 safe
//var MIN_COIN = common.Big0

type Safe3Storage struct {
    dataPath     string
    solcPath     string
    contractPath string
    storagePath  string
    ownerAddr    string
    isStorage    bool
}

func NewSafe3Storage(tool *types.Tool) *Safe3Storage {
    return &Safe3Storage{
        dataPath:     tool.GetDataPath(),
        solcPath:     tool.GetSolcPath(),
        contractPath: tool.GetContractPath(),
        storagePath:  tool.GetSafe3StoragePath(),
        ownerAddr:    tool.GetOwnerAddress(),
        isStorage:    utils.IsStorage(),
    }
}

func (s *Safe3Storage) Generate(alloc *types.GenesisAlloc) {
    utils.Compile(s.solcPath, s.contractPath, "Safe3.sol")

    totalAmount := big.NewInt(0)
    lockedAmounts := s.loadLockedInfos(totalAmount)
    specialAmounts := s.loadSpecialInfos(totalAmount)
    s.loadBalance(lockedAmounts, specialAmounts, totalAmount)

    contractNames := [2]string{"TransparentUpgradeableProxy", "Safe3"}
    contractAddrs := [2]string{"0x0000000000000000000000000000000000001090", "0x0000000000000000000000000000000000001091"}

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
            account.Balance = totalAmount.String()
            account.Storage = make(map[common.Hash]common.Hash)
            account.Storage[common.BigToHash(big.NewInt(0))] = common.BigToHash(big.NewInt(1))
            account.Storage[common.BigToHash(big.NewInt(0x33))] = common.HexToHash(s.ownerAddr)
            account.Storage[common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")] = common.HexToHash(contractAddrs[1])
            account.Storage[common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")] = common.HexToHash(ProxyAdminAddr)

            // specialKeyIDs
            s.buildSpecialKeyIDs(&account, specialAmounts)

            // specials
            s.buildSpecials(&account, specialAmounts)
        }
        (*alloc)[common.HexToAddress(contractAddrs[i])] = account
    }

    os.RemoveAll(filepath.Join(s.contractPath, "temp"))

    if s.isStorage {
        os.MkdirAll(s.storagePath, 0755)
        fileName := filepath.Join(s.storagePath, "storage_list.go")
        f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
        if err != nil {
            panic(err)
        }
        defer f.Close()

        temp := fmt.Sprintf("package %s\n\n", filepath.Base(s.storagePath))
        temp += "var StorageList = []string {\n"
        for _, key := range storageList {
            temp += "    " + key + ",\n"
        }
        temp += "}\n"

        if _, err = fmt.Fprintf(f, "%s", temp); err != nil {
            panic(err)
        }
    }
}

func (s *Safe3Storage) unzip(zipPath string) {
    dst := filepath.Dir(zipPath)
    archive, err := zip.OpenReader(zipPath)
    if err != nil {
        panic(err)
    }
    defer archive.Close()

    for _, f := range archive.File {
        filePath := filepath.Join(dst, f.Name)
        dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
        if err != nil {
            panic(err)
        }

        fileInArchive, err := f.Open()
        if err != nil {
            panic(err)
        }

        if _, err := io.Copy(dstFile, fileInArchive); err != nil {
            panic(err)
        }

        dstFile.Close()
        fileInArchive.Close()
    }
}

func (s *Safe3Storage) loadBalance(lockedAmounts map[string]*big.Int, specialAmounts map[string]*big.Int, totalAmount *big.Int) map[string]*big.Int {
    s.unzip(filepath.Join(s.dataPath, "safe3", "balanceaddress.zip"))
    file, err := os.Open(filepath.Join(s.dataPath, "safe3", "balanceaddress.csv"))
    if err != nil {
        panic(err)
    }

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
        if amount.Cmp(MIN_COIN) <= 0 {
            continue
        }

        if specialAmounts[addr] != nil {
            continue
        }

        lockedAmount, _ := new(big.Int).SetString(temps[3], 10)
        if lockedAmounts[addr] != nil && lockedAmount.Cmp(lockedAmounts[addr]) <= 0 {
            lockedAmount = lockedAmounts[addr]
        }

        temp := big.NewInt(0).Sub(amount, lockedAmount)
        if temp.Cmp(MIN_COIN) <= 0 {
            continue
        }
        totalAmount.Add(totalAmount, temp)
        availableAmounts[addr] = temp
    }

    file.Close()
    os.Remove(filepath.Join(s.dataPath, "safe3", "balanceaddress.csv"))

    if s.isStorage {
        // split availables
        // keyIDs
        if len(availableAmounts) > 0 {
            // size
            storageKey := common.BigToHash(big.NewInt(101))
            storageValue := common.BigToHash(big.NewInt(int64(len(availableAmounts))))
            pairs[storageKey] = storageValue
            if len(pairs) >= maxNum {
                s.save()
            }
            // items
            subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(101))
            i := int64(0)
            for addr := range availableAmounts {
                curKey := big.NewInt(0).Add(subKey, big.NewInt(i))
                keyID := getKeyIDFromAddress(addr)
                subStorageKeys, subStorageValues := utils.GetStorage4Bytes(curKey, keyID)
                for k := range subStorageKeys {
                    pairs[subStorageKeys[k]] = subStorageValues[k]
                    if len(pairs) >= maxNum {
                        s.save()
                    }
                }
                i++
            }
        }
        // availables
        for addr, amount := range availableAmounts {
            curKey := big.NewInt(0).SetBytes(utils.Keccak256_uint_bytes(102, getKeyIDFromAddress(addr)))
            storageKey, storageValue := utils.GetStorage4Int(curKey, amount)
            pairs[storageKey] = storageValue
            if len(pairs) >= maxNum {
                s.save()
            }
        }
        if len(pairs) > 0 {
            s.save()
        }
    }
    fmt.Printf("available address: %d\n", len(availableAmounts))
    return availableAmounts
}

func (s *Safe3Storage) loadSpecialInfos(totalAmount *big.Int) map[string]*big.Int {
    file, err := os.Open(filepath.Join(s.dataPath, "safe3", "specialaddress.csv"))
    if err != nil {
        panic(err)
    }
    defer file.Close()

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
        if amount.Cmp(MIN_COIN) <= 0 {
            continue
        }
        totalAmount.Add(totalAmount, amount)
        specialAmounts[addr] = amount
    }
    return specialAmounts
}

func (s *Safe3Storage) loadMNs() map[string]string {
    jsonFile, err := os.Open(filepath.Join(s.dataPath, "safe3", "masternodes.csv"))
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

func (s *Safe3Storage) loadLockedInfos(totalAmount *big.Int) map[string]*big.Int {
    masternodes := s.loadMNs()

    s.unzip(filepath.Join(s.dataPath, "safe3", "lockedaddresses.zip"))
    file, err := os.Open(filepath.Join(s.dataPath, "safe3", "lockedaddresses.csv"))
    if err != nil {
        panic(err)
    }

    lockedInfos := make(map[string][]types.LockedData)
    lockedAmounts := make(map[string]*big.Int)
    lockedNum := int64(0)

    BTC_COIN := new(big.Float).SetInt(big.NewInt(100000000))
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
        amt.Mul(amt, BTC_COIN)
        amount, _ := amt.Int(nil)
        if amount.Cmp(MIN_COIN) <= 0 {
            continue
        }
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
            if amount.Cmp(MIN_COIN) <= 0 {
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
        totalAmount.Add(totalAmount, amount)
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

    file.Close()
    os.Remove(filepath.Join(s.dataPath, "safe3", "lockedaddresses.csv"))

    if s.isStorage {
        // split locks
        // lockedNum
        if lockedNum > 0 {
            storageKey, storageValue := utils.GetStorage4Int(big.NewInt(103), big.NewInt(lockedNum))
            pairs[storageKey] = storageValue
            if len(pairs) >= maxNum {
                s.save()
            }
        }
        // lockedKeyIDs
        if len(lockedInfos) > 0 {
            // size
            storageKey := common.BigToHash(big.NewInt(104))
            storageValue := common.BigToHash(big.NewInt(int64(len(lockedInfos))))
            pairs[storageKey] = storageValue
            if len(pairs) >= maxNum {
                s.save()
            }
            // items
            subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(104))
            i := int64(0)
            for addr := range lockedInfos {
                curKey := big.NewInt(0).Add(subKey, big.NewInt(i))
                keyID := getKeyIDFromAddress(addr)
                subStorageKeys, subStorageValues := utils.GetStorage4Bytes(curKey, keyID)
                for k := range subStorageKeys {
                    pairs[subStorageKeys[k]] = subStorageValues[k]
                    if len(pairs) >= maxNum {
                        s.save()
                    }
                }
                i++
            }
        }
        // locks
        for addr, list := range lockedInfos {
            // size
            curKey := big.NewInt(0).SetBytes(utils.Keccak256_uint_bytes(105, getKeyIDFromAddress(addr)))
            storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(int64(len(list))))
            pairs[storageKey] = storageValue
            if len(pairs) >= maxNum {
                s.save()
            }

            curKey = big.NewInt(0).SetBytes(utils.Keccak256_bytes32(common.BigToHash(curKey).Hex()))
            curKey.Sub(curKey, common.Big1)
            for _, info := range list {
                // amount(8) + remainLockHeight(4) + lockDay(2) + isMn(1) + redeemHeight(4)
                curKey.Add(curKey, common.Big1)
                storageKey = common.BigToHash(curKey)
                amountStorageValue := common.BigToHash(info.Amount)                     // amount: 8 bytes
                remainLockHeightStorageValue := common.BigToHash(info.RemainLockHeight) // remainLockHeight: 4 bytes
                lockDayStorageValue := common.BigToHash(info.LockDay)                   // lockDay: 2 bytes
                isMNStorageValue := common.Hash{}                                       // isMN: 1 bytes
                if info.IsMN {
                    isMNStorageValue = common.BigToHash(big.NewInt(1))
                } else {
                    isMNStorageValue = common.BigToHash(big.NewInt(0))
                }

                storageValue = common.Hash{}
                offset := 0
                for i := 0; i < 8; i++ {
                    storageValue[31-i-offset] = amountStorageValue[31-i]
                }
                offset += 8
                for i := 0; i < 4; i++ {
                    storageValue[31-i-offset] = remainLockHeightStorageValue[31-i]
                }
                offset += 4
                for i := 0; i < 2; i++ {
                    storageValue[31-i-offset] = lockDayStorageValue[31-i]
                }
                offset += 2
                storageValue[31-offset] = isMNStorageValue[31]

                pairs[storageKey] = storageValue
                if len(pairs) >= maxNum {
                    s.save()
                }
                curKey.Add(curKey, common.Big1)
            }
        }
    }
    fmt.Printf("total locked number: %d, total locked address: %d\n", lockedNum, len(lockedAmounts))
    return lockedAmounts
}

func (s *Safe3Storage) buildKeyIDs(account *types.GenesisAccount, availableAmounts map[string]*big.Int) {
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

func (s *Safe3Storage) buildAvailables(account *types.GenesisAccount, availableAmounts map[string]*big.Int) {
    var curKey *big.Int
    for addr, amount := range availableAmounts {
        curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_bytes(102, getKeyIDFromAddress(addr)))
        s.calcAmount(account, amount, &curKey)
    }
}

func (s *Safe3Storage) calcAmount(account *types.GenesisAccount, amount *big.Int, curKey **big.Int) {
    storageKey, storageValue := utils.GetStorage4Int(*curKey, amount)
    account.Storage[storageKey] = storageValue
}

func (s *Safe3Storage) buildLockedNum(account *types.GenesisAccount, lockedNum int64) {
    curKey := big.NewInt(103)
    storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(lockedNum))
    account.Storage[storageKey] = storageValue
}

func (s *Safe3Storage) buildLockedKeyIDs(account *types.GenesisAccount, infos map[string][]types.LockedData) {
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

func (s *Safe3Storage) buildLocks(account *types.GenesisAccount, infos map[string][]types.LockedData) {
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
            s.calcTxid(account, info, &curKey)
            s.calcPart(account, info, &curKey)
            curKey = curKey.Add(curKey, big.NewInt(1))
        }
    }
}

func (s *Safe3Storage) calcTxid(account *types.GenesisAccount, info types.LockedData, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue := utils.GetStorage4Bytes32(*curKey, info.Txid)
    account.Storage[storageKey] = storageValue
}

func (s *Safe3Storage) calcPart(account *types.GenesisAccount, info types.LockedData, curKey **big.Int) {
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

func (s *Safe3Storage) buildSpecialKeyIDs(account *types.GenesisAccount, specialAmounts map[string]*big.Int) {
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

func (s *Safe3Storage) buildSpecials(account *types.GenesisAccount, specialAmounts map[string]*big.Int) {
    var curKey *big.Int
    for addr, amount := range specialAmounts {
        curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_bytes(107, getKeyIDFromAddress(addr)))
        s.calcAmount3(account, amount, &curKey)
    }
}

func (s *Safe3Storage) calcAmount3(account *types.GenesisAccount, amount *big.Int, curKey **big.Int) {
    storageKey, storageValue := utils.GetStorage4Int(*curKey, amount)
    account.Storage[storageKey] = storageValue
}

func (s *Safe3Storage) save() {
    os.MkdirAll(s.storagePath, 0755)

    fileIndex++
    varName := "Storage" + strconv.Itoa(fileIndex)
    filePath := filepath.Join(s.storagePath, "storage"+strconv.Itoa(fileIndex)+".go")
    f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    b, _ := json.Marshal(pairs)

    var buf bytes.Buffer
    gz := gzip.NewWriter(&buf)
    if _, err := gz.Write(b); err != nil {
        panic(err)
    }
    if err := gz.Close(); err != nil {
        panic(err)
    }

    if _, err := fmt.Fprintf(f, "package %s\n\nvar %s = %q\n", filepath.Base(s.storagePath), varName, buf.Bytes()); err != nil {
        panic(err)
    }
    storageList = append(storageList, varName)

    pairs = make(map[common.Hash]common.Hash)
}

func getKeyIDFromAddress(addr string) []byte {
    return utils.Base58Decoding(addr)
}

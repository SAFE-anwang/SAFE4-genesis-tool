package contracts

import (
    "encoding/json"
    "github.com/safe/SAFE4-genesis-tool/common"
    "github.com/safe/SAFE4-genesis-tool/types"
    "github.com/safe/SAFE4-genesis-tool/utils"
    "io/ioutil"
    "math/big"
    "os"
    "path/filepath"
)

type AccountManagerStorage struct {
    dataPath     string
    solcPath     string
    contractPath string
    ownerAddr    string
}

func NewAccountManagerStorage(tool *types.Tool) *AccountManagerStorage {
    return &AccountManagerStorage{
        dataPath:     tool.GetDataPath(),
        solcPath:     tool.GetSolcPath(),
        contractPath: tool.GetContractPath(),
        ownerAddr:    tool.GetOwnerAddress(),
    }
}

func (s *AccountManagerStorage) Generate(alloc *types.GenesisAlloc) {
    utils.Compile(s.solcPath, s.contractPath, "AccountManager.sol")

    masternodes := s.loadMasterNode()
    supernodes := s.loadSuperNode()

    totalAmount := big.NewInt(0)
    addr2amounts := make(map[common.Address][]*big.Int)
    addr2addrs := make(map[common.Address][]common.Address)
    var addrs []common.Address
    recordNo := 0
    for _, masternode := range *masternodes {
        addr := masternode.Founders[0].Addr
        amount := masternode.Founders[0].Amount
        totalAmount = totalAmount.Add(totalAmount, amount)
        if len(addr2amounts[addr]) == 0 {
            addrs = append(addrs, addr)
        }
        addr2amounts[addr] = append(addr2amounts[addr], amount)
        addr2addrs[addr] = append(addr2addrs[addr], masternode.Addr)
        recordNo++
    }
    for _, supernode := range *supernodes {
        addr := supernode.Founders[0].Addr
        amount := supernode.Founders[0].Amount
        totalAmount = totalAmount.Add(totalAmount, amount)
        if len(addr2amounts[addr]) == 0 {
            addrs = append(addrs, addr)
        }
        addr2amounts[addr] = append(addr2amounts[addr], amount)
        addr2addrs[addr] = append(addr2addrs[addr], supernode.Addr)
        recordNo++
    }

    contractNames := [2]string{"TransparentUpgradeableProxy", "AccountManager"}
    contractAddrs := [2]string{"0x0000000000000000000000000000000000001010", "0x0000000000000000000000000000000000001011"}

    for i := range contractNames {
        code, err := os.ReadFile(filepath.Join(s.contractPath, "temp", contractNames[i]+".bin-runtime"))
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

            // record_no
            s.buildRecordNo(&account, recordNo)

            // addr2records
            s.buildAddr2records(&account, addrs, addr2amounts)

            // id2index
            s.buildID2index(&account, addrs, addr2amounts)

            // id2addr
            s.buildID2addr(&account, addrs, addr2amounts)

            // id2useinfo
            s.buildID2useInfo(&account, addrs, addr2addrs)
        }
        (*alloc)[common.HexToAddress(contractAddrs[i])] = account
    }

    os.RemoveAll(filepath.Join(s.contractPath, "temp"))
}

func (s *AccountManagerStorage) loadMasterNode() *[]types.MasterNodeInfo {
    jsonFile, err := os.Open(filepath.Join(s.dataPath, "MasterNode.info"))
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

func (s *AccountManagerStorage) loadSuperNode() *[]types.SuperNodeInfo {
    jsonFile, err := os.Open(filepath.Join(s.dataPath, "SuperNode.info"))
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

func (s *AccountManagerStorage) buildRecordNo(account *types.GenesisAccount, count int) {
    curKey := big.NewInt(102)
    storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(int64(count)))
    account.Storage[storageKey] = storageValue
}

func (s *AccountManagerStorage) buildAddr2records(account *types.GenesisAccount, addrs []common.Address, addr2amounts map[common.Address][]*big.Int) {
    var curKey *big.Int
    var storageKey, storageValue common.Hash

    id := int64(1)
    for _, addr := range addrs {
        // size
        curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_address(103, addr))
        storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(int64(len(addr2amounts[addr]))))
        account.Storage[storageKey] = storageValue

        itemKey := big.NewInt(0).SetBytes(utils.Keccak256_bytes32(common.BigToHash(curKey).Hex()))
        for _, amount := range addr2amounts[addr] {
            // id
            storageKey, storageValue = utils.GetStorage4Int(itemKey, big.NewInt(id))
            account.Storage[storageKey] = storageValue
            id++
            // addr
            itemKey = big.NewInt(0).Add(itemKey, big.NewInt(1))
            storageKey, storageValue = utils.GetStorage4Addr(itemKey, addr)
            account.Storage[storageKey] = storageValue
            // amount
            itemKey = big.NewInt(0).Add(itemKey, big.NewInt(1))
            storageKey, storageValue = utils.GetStorage4Int(itemKey, amount)
            account.Storage[storageKey] = storageValue
            // lockDay
            itemKey = big.NewInt(0).Add(itemKey, big.NewInt(1))
            storageKey, storageValue = utils.GetStorage4Int(itemKey, big.NewInt(720))
            account.Storage[storageKey] = storageValue
            // startHeight
            itemKey = big.NewInt(0).Add(itemKey, big.NewInt(1))
            storageKey, storageValue = utils.GetStorage4Int(itemKey, big.NewInt(0))
            account.Storage[storageKey] = storageValue
            // unlockHeight
            itemKey = big.NewInt(0).Add(itemKey, big.NewInt(1))
            storageKey, storageValue = utils.GetStorage4Int(itemKey, big.NewInt(720*24*3600/30))
            account.Storage[storageKey] = storageValue

            itemKey = big.NewInt(0).Add(itemKey, big.NewInt(1))
        }
    }
}

func (s *AccountManagerStorage) buildID2index(account *types.GenesisAccount, addrs []common.Address, addr2amounts map[common.Address][]*big.Int) {
    var curKey *big.Int
    var storageKey, storageValue common.Hash

    id := int64(1)
    for _, addr := range addrs {
        for i, _ := range addr2amounts[addr] {
            curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(104, id))
            storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(int64(i)))
            account.Storage[storageKey] = storageValue
            id++
        }
    }
}

func (s *AccountManagerStorage) buildID2addr(account *types.GenesisAccount, addrs []common.Address, addr2amounts map[common.Address][]*big.Int) {
    var curKey *big.Int
    var storageKey, storageValue common.Hash

    id := int64(1)
    for _, addr := range addrs {
        for _, _ = range addr2amounts[addr] {
            curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(105, id))
            storageKey, storageValue = utils.GetStorage4Addr(curKey, addr)
            account.Storage[storageKey] = storageValue
            id++
        }
    }
}

func (s *AccountManagerStorage) buildID2useInfo(account *types.GenesisAccount, addrs []common.Address, addr2addrs map[common.Address][]common.Address) {
    var curKey *big.Int
    var storageKey, storageValue common.Hash

    id := int64(1)
    for _, addr := range addrs {
        for _, targetAddr := range addr2addrs[addr] {
            // frozenAddr
            curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(106, id))
            storageKey, storageValue = utils.GetStorage4Addr(curKey, targetAddr)
            account.Storage[storageKey] = storageValue
            id++
            // freezeHeight
            curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
            storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(0))
            account.Storage[storageKey] = storageValue
            // unfreezeHeight
            curKey = big.NewInt(0).Add(curKey, big.NewInt(1))
            storageKey, storageValue = utils.GetStorage4Int(curKey, big.NewInt(720*24*3600/30))
            account.Storage[storageKey] = storageValue
        }
    }
}

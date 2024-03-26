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

type SuperNodeStorageStorage struct {
    dataPath     string
    solcPath     string
    contractPath string
    ownerAddr    string
}

func NewSuperNodeStorageStorage(tool *types.Tool) *SuperNodeStorageStorage {
    return &SuperNodeStorageStorage{
        dataPath:     tool.GetDataPath(),
        solcPath:     tool.GetSolcPath(),
        contractPath: tool.GetContractPath(),
        ownerAddr:    tool.GetOwnerAddress(),
    }
}

func (s *SuperNodeStorageStorage) Generate(alloc *types.GenesisAlloc) {
    utils.Compile(s.solcPath, s.contractPath, "SuperNodeStorage.sol")

    supernodes := s.load()

    contractNames := [2]string{"TransparentUpgradeableProxy", "SuperNodeStorage"}
    contractAddrs := [2]string{"0x0000000000000000000000000000000000001030", "0x0000000000000000000000000000000000001031"}

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
            account.Storage = make(map[common.Hash]common.Hash)
            account.Storage[common.BigToHash(big.NewInt(0))] = common.BigToHash(big.NewInt(1))
            account.Storage[common.BigToHash(big.NewInt(0x33))] = common.HexToHash(s.ownerAddr)
            account.Storage[common.HexToHash("0x360894a13ba1a3210667c828492db98dca3e2076cc3735a920a3ca505d382bbc")] = common.HexToHash(contractAddrs[1])
            account.Storage[common.HexToHash("0xb53127684a568b3173ae13b9f8a6016e243e63b6e8ee1178d6a717850b5d6103")] = common.HexToHash(ProxyAdminAddr)

            // no
            s.buildNo(&account, supernodes)

            // addr2info
            s.buildAddr2Info(&account, supernodes)

            // ids
            s.buildIDs(&account, supernodes)

            // id2addr
            s.buildID2Addr(&account, supernodes)

            // name2addr
            s.buildName2Addr(&account, supernodes)

            // enode2addr
            s.buildEnode2Addr(&account, supernodes)
        }
        (*alloc)[common.HexToAddress(contractAddrs[i])] = account
    }

    os.RemoveAll(filepath.Join(s.contractPath, "temp"))
}

func (s *SuperNodeStorageStorage) load() *[]types.SuperNodeInfo {
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

func (s *SuperNodeStorageStorage) buildNo(account *types.GenesisAccount, supernodes *[]types.SuperNodeInfo) {
    curKey := big.NewInt(101)
    storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(int64(len(*supernodes))))
    account.Storage[storageKey] = storageValue
}

func (s *SuperNodeStorageStorage) buildAddr2Info(account *types.GenesisAccount, supernodes *[]types.SuperNodeInfo) {
    var curKey *big.Int
    for _, supernode := range *supernodes {
        s.calcId(account, supernode, &curKey)
        s.calcName(account, supernode, &curKey)
        s.calcAddr(account, supernode, &curKey)
        s.calcCreator(account, supernode, &curKey)
        s.calcEnode(account, supernode, &curKey)
        s.calcDesc(account, supernode, &curKey)
        s.calcIsOfficial(account, supernode, &curKey)
        s.calcState(account, supernode, &curKey)
        s.calcFounders(account, supernode, &curKey)
        s.calcIncentivePlan(account, supernode, &curKey)
    }
}

func (s *SuperNodeStorageStorage) calcId(account *types.GenesisAccount, supernode types.SuperNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_address(102, supernode.Addr))
    storageKey, storageValue := utils.GetStorage4Int(*curKey, supernode.Id)
    account.Storage[storageKey] = storageValue
}

func (s *SuperNodeStorageStorage) calcName(account *types.GenesisAccount, supernode types.SuperNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKeys, storageValues := utils.GetStorage4String(*curKey, supernode.Name)
    for i := range storageKeys {
        account.Storage[storageKeys[i]] = storageValues[i]
    }
}

func (s *SuperNodeStorageStorage) calcAddr(account *types.GenesisAccount, supernode types.SuperNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue := utils.GetStorage4Addr(*curKey, supernode.Addr)
    account.Storage[storageKey] = storageValue
}

func (s *SuperNodeStorageStorage) calcCreator(account *types.GenesisAccount, supernode types.SuperNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue := utils.GetStorage4Addr(*curKey, supernode.Creator)
    account.Storage[storageKey] = storageValue
}

func (s *SuperNodeStorageStorage) calcEnode(account *types.GenesisAccount, supernode types.SuperNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKeys, storageValues := utils.GetStorage4String(*curKey, supernode.Enode)
    for i := range storageKeys {
        account.Storage[storageKeys[i]] = storageValues[i]
    }
}

func (s *SuperNodeStorageStorage) calcDesc(account *types.GenesisAccount, supernode types.SuperNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKeys, storageValues := utils.GetStorage4String(*curKey, supernode.Description)
    for i := range storageKeys {
        account.Storage[storageKeys[i]] = storageValues[i]
    }
}

func (s *SuperNodeStorageStorage) calcIsOfficial(account *types.GenesisAccount, supernode types.SuperNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue := utils.GetStorage4Bool(*curKey, supernode.IsOfficial)
    account.Storage[storageKey] = storageValue
}

func (s *SuperNodeStorageStorage) calcState(account *types.GenesisAccount, supernode types.SuperNodeInfo, curKey **big.Int) {
    var storageKey, storageValue common.Hash
    // state
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue = utils.GetStorage4Int(*curKey, supernode.State)
    account.Storage[storageKey] = storageValue
}

func (s *SuperNodeStorageStorage) calcFounders(account *types.GenesisAccount, supernode types.SuperNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey := common.BigToHash(*curKey)
    storageValue := common.BigToHash(big.NewInt(int64(len(supernode.Founders))))
    account.Storage[storageKey] = storageValue

    subKey := big.NewInt(0).SetBytes(utils.Keccak256_bytes32(common.BigToHash(*curKey).Hex()))
    var subStorageKey, subStorageValue common.Hash
    for _, founder := range supernode.Founders {
        // lockID
        subStorageKey, subStorageValue = utils.GetStorage4Int(subKey, founder.LockID)
        account.Storage[subStorageKey] = subStorageValue
        subKey = subKey.Add(subKey, big.NewInt(1))

        // addr
        subStorageKey, subStorageValue = utils.GetStorage4Addr(subKey, founder.Addr)
        account.Storage[subStorageKey] = subStorageValue
        subKey = subKey.Add(subKey, big.NewInt(1))

        // amount
        subStorageKey, subStorageValue = utils.GetStorage4Int(subKey, founder.Amount)
        account.Storage[subStorageKey] = subStorageValue
        subKey = subKey.Add(subKey, big.NewInt(1))

        // height
        subStorageKey, subStorageValue = utils.GetStorage4Int(subKey, founder.Height)
        account.Storage[subStorageKey] = subStorageValue
        subKey = subKey.Add(subKey, big.NewInt(1))
    }
}

func (s *SuperNodeStorageStorage) calcIncentivePlan(account *types.GenesisAccount, supernode types.SuperNodeInfo, curKey **big.Int) {
    var storageKey, storageValue common.Hash
    // creator
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue = utils.GetStorage4Int(*curKey, supernode.IncentivePlan.Creator)
    account.Storage[storageKey] = storageValue

    // partner
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue = utils.GetStorage4Int(*curKey, supernode.IncentivePlan.Partner)
    account.Storage[storageKey] = storageValue

    // voter
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue = utils.GetStorage4Int(*curKey, supernode.IncentivePlan.Voter)
    account.Storage[storageKey] = storageValue
}

func (s *SuperNodeStorageStorage) buildIDs(account *types.GenesisAccount, supernodes *[]types.SuperNodeInfo) {
    storageKey := common.BigToHash(big.NewInt(103))
    storageValue := common.BigToHash(big.NewInt(int64(len(*supernodes))))
    account.Storage[storageKey] = storageValue

    subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(103))
    for i, supernode := range *supernodes {
        curKey := big.NewInt(0).Add(subKey, big.NewInt(int64(i)))
        subStorageKey, subStorageValue := utils.GetStorage4Int(curKey, supernode.Id)
        account.Storage[subStorageKey] = subStorageValue
    }
}

func (s *SuperNodeStorageStorage) buildID2Addr(account *types.GenesisAccount, supernodes *[]types.SuperNodeInfo) {
    for _, supernode := range *supernodes {
        curKey := big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(104, supernode.Id.Int64()))
        storageKey, storageValue := utils.GetStorage4Addr(curKey, supernode.Addr)
        account.Storage[storageKey] = storageValue
    }
}

func (s *SuperNodeStorageStorage) buildName2Addr(account *types.GenesisAccount, supernodes *[]types.SuperNodeInfo) {
    for _, supernode := range *supernodes {
        curKey := big.NewInt(0).SetBytes(utils.Keccak256_uint_string(105, supernode.Name))
        storageKey, storageValue := utils.GetStorage4Addr(curKey, supernode.Addr)
        account.Storage[storageKey] = storageValue
    }
}

func (s *SuperNodeStorageStorage) buildEnode2Addr(account *types.GenesisAccount, supernodes *[]types.SuperNodeInfo) {
    for _, supernode := range *supernodes {
        curKey := big.NewInt(0).SetBytes(utils.Keccak256_uint_string(106, supernode.Enode))
        storageKey, storageValue := utils.GetStorage4Addr(curKey, supernode.Addr)
        account.Storage[storageKey] = storageValue
    }
}

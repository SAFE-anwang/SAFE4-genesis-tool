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

type MasterNodeStorageStorage struct {
    dataPath     string
    solcPath     string
    contractPath string
    ownerAddr    string
}

func NewMasterNodeStorageStorage(tool *types.Tool) *MasterNodeStorageStorage {
    return &MasterNodeStorageStorage{
        dataPath:     tool.GetDataPath(),
        solcPath:     tool.GetSolcPath(),
        contractPath: tool.GetContractPath(),
        ownerAddr:    tool.GetOwnerAddress(),
    }
}

func (s *MasterNodeStorageStorage) Generate(alloc *types.GenesisAlloc) {
    utils.Compile(s.solcPath, s.contractPath, "MasterNodeStorage.sol")

    masternodes := s.load()

    contractNames := [2]string{"TransparentUpgradeableProxy", "MasterNodeStorage"}
    contractAddrs := [2]string{"0x0000000000000000000000000000000000001020", "0x0000000000000000000000000000000000001021"}

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
            s.buildNo(&account, masternodes)

            // addr2info
            s.buildAddr2Info(&account, masternodes)

            // ids
            s.buildIDs(&account, masternodes)

            // id2addr
            s.buildID2Addr(&account, masternodes)

            // enode2addr
            s.buildEnode2Addr(&account, masternodes)
        }
        (*alloc)[common.HexToAddress(contractAddrs[i])] = account
    }

    os.RemoveAll(filepath.Join(s.contractPath, "temp"))
}

func (s *MasterNodeStorageStorage) load() *[]types.MasterNodeInfo {
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

func (s *MasterNodeStorageStorage) buildNo(account *types.GenesisAccount, masternodes *[]types.MasterNodeInfo) {
    curKey := big.NewInt(101)
    storageKey, storageValue := utils.GetStorage4Int(curKey, big.NewInt(int64(len(*masternodes))))
    account.Storage[storageKey] = storageValue
}

func (s *MasterNodeStorageStorage) buildAddr2Info(account *types.GenesisAccount, masternodes *[]types.MasterNodeInfo) {
    var curKey *big.Int
    for _, masternode := range *masternodes {
        s.calcId(account, masternode, &curKey)
        s.calcAddr(account, masternode, &curKey)
        s.calcCreator(account, masternode, &curKey)
        s.calcEnode(account, masternode, &curKey)
        s.calcDesc(account, masternode, &curKey)
        s.calcIsOfficial(account, masternode, &curKey)
        s.calcState(account, masternode, &curKey)
        s.calcFounders(account, masternode, &curKey)
        s.calcIncentivePlan(account, masternode, &curKey)
        s.calcIsUnion(account, masternode, &curKey)
    }
}

func (s *MasterNodeStorageStorage) calcId(account *types.GenesisAccount, masternode types.MasterNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).SetBytes(utils.Keccak256_uint_address(102, masternode.Addr))
    storageKey, storageValue := utils.GetStorage4Int(*curKey, masternode.Id)
    account.Storage[storageKey] = storageValue
}

func (s *MasterNodeStorageStorage) calcAddr(account *types.GenesisAccount, masternode types.MasterNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue := utils.GetStorage4Addr(*curKey, masternode.Addr)
    account.Storage[storageKey] = storageValue
}

func (s *MasterNodeStorageStorage) calcCreator(account *types.GenesisAccount, masternode types.MasterNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue := utils.GetStorage4Addr(*curKey, masternode.Creator)
    account.Storage[storageKey] = storageValue
}

func (s *MasterNodeStorageStorage) calcEnode(account *types.GenesisAccount, masternode types.MasterNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKeys, storageValues := utils.GetStorage4String(*curKey, masternode.Enode)
    for i := range storageKeys {
        account.Storage[storageKeys[i]] = storageValues[i]
    }
}

func (s *MasterNodeStorageStorage) calcDesc(account *types.GenesisAccount, masternode types.MasterNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKeys, storageValues := utils.GetStorage4String(*curKey, masternode.Description)
    for i := range storageKeys {
        account.Storage[storageKeys[i]] = storageValues[i]
    }
}

func (s *MasterNodeStorageStorage) calcIsOfficial(account *types.GenesisAccount, masternode types.MasterNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue := utils.GetStorage4Bool(*curKey, masternode.IsOfficial)
    account.Storage[storageKey] = storageValue
}

func (s *MasterNodeStorageStorage) calcState(account *types.GenesisAccount, masternode types.MasterNodeInfo, curKey **big.Int) {
    var storageKey, storageValue common.Hash
    // state
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue = utils.GetStorage4Int(*curKey, masternode.State)
    account.Storage[storageKey] = storageValue
}

func (s *MasterNodeStorageStorage) calcFounders(account *types.GenesisAccount, masternode types.MasterNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey := common.BigToHash(*curKey)
    storageValue := common.BigToHash(big.NewInt(int64(len(masternode.Founders))))
    account.Storage[storageKey] = storageValue

    subKey := big.NewInt(0).SetBytes(utils.Keccak256_bytes32(common.BigToHash(*curKey).Hex()))
    var subStorageKey, subStorageValue common.Hash
    for _, founder := range masternode.Founders {
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

func (s *MasterNodeStorageStorage) calcIncentivePlan(account *types.GenesisAccount, masternode types.MasterNodeInfo, curKey **big.Int) {
    var storageKey, storageValue common.Hash
    // creator
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue = utils.GetStorage4Int(*curKey, masternode.IncentivePlan.Creator)
    account.Storage[storageKey] = storageValue

    // partner
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue = utils.GetStorage4Int(*curKey, masternode.IncentivePlan.Partner)
    account.Storage[storageKey] = storageValue

    // voter
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue = utils.GetStorage4Int(*curKey, masternode.IncentivePlan.Voter)
    account.Storage[storageKey] = storageValue
}

func (s *MasterNodeStorageStorage) calcIsUnion(account *types.GenesisAccount, masternode types.MasterNodeInfo, curKey **big.Int) {
    *curKey = big.NewInt(0).Add(*curKey, big.NewInt(1))
    storageKey, storageValue := utils.GetStorage4Bool(*curKey, masternode.IsUnion)
    account.Storage[storageKey] = storageValue
}

func (s *MasterNodeStorageStorage) buildIDs(account *types.GenesisAccount, masternodes *[]types.MasterNodeInfo) {
    storageKey := common.BigToHash(big.NewInt(103))
    storageValue := common.BigToHash(big.NewInt(int64(len(*masternodes))))
    account.Storage[storageKey] = storageValue

    subKey := big.NewInt(0).SetBytes(utils.Keccak256_uint(103))
    for i, masternode := range *masternodes {
        curKey := big.NewInt(0).Add(subKey, big.NewInt(int64(i)))
        subStorageKey, subStorageValue := utils.GetStorage4Int(curKey, masternode.Id)
        account.Storage[subStorageKey] = subStorageValue
    }
}

func (s *MasterNodeStorageStorage) buildID2Addr(account *types.GenesisAccount, masternodes *[]types.MasterNodeInfo) {
    for _, masternode := range *masternodes {
        curKey := big.NewInt(0).SetBytes(utils.Keccak256_uint_uint(104, masternode.Id.Int64()))
        storageKey, storageValue := utils.GetStorage4Addr(curKey, masternode.Addr)
        account.Storage[storageKey] = storageValue
    }
}

func (s *MasterNodeStorageStorage) buildEnode2Addr(account *types.GenesisAccount, masternodes *[]types.MasterNodeInfo) {
    for _, masternode := range *masternodes {
        curKey := big.NewInt(0).SetBytes(utils.Keccak256_uint_string(105, masternode.Enode))
        storageKey, storageValue := utils.GetStorage4Addr(curKey, masternode.Addr)
        account.Storage[storageKey] = storageValue
    }
}

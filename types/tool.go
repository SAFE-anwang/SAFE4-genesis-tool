package types

import (
    "encoding/json"
    "fmt"
    js "github.com/dop251/goja"
    "github.com/safe/SAFE4-genesis-tool/common"
    "github.com/safe/SAFE4-genesis-tool/params"
    "github.com/safe/SAFE4-genesis-tool/utils"
    "io/ioutil"
    "math/big"
    "os"
    "path/filepath"
    "strings"
)

type Tool struct {
    netType   int
    workPath  string
    depPath   string
    isStorage bool
    genesis   Genesis
}

func NewTool() (*Tool, error) {
    ex, err := os.Executable()
    if err != nil {
        return nil, err
    }
    return &Tool{
        netType:   utils.GetNetType(),
        workPath:  filepath.Dir(ex),
        depPath:   filepath.Join(filepath.Dir(ex), "..", "deps"),
        isStorage: utils.IsStorage(),
    }, err
}

func (t *Tool) GetOwnerAddress() string {
    if t.netType == 0 {
        return "0xac110c0f70867f77d9d230e377043f52480a0b7d"
    } else if t.netType == 1 {
        return "0xac110c0f70867f77d9d230e377043f52480a0b7d"
    } else {
        return "0x80d8b8f308770ce14252173abb00075cc9082d03"
    }
}

func (t *Tool) GetOwnerBalance() *big.Int {
    balance, _ := new(big.Int).SetString("100000000000000000000000", 10)
    if t.netType == 0 {
        return balance
    } else if t.netType == 1 {
        return big.NewInt(1).Mul(balance, big.NewInt(10000000))
    } else {
        return big.NewInt(1).Mul(balance, big.NewInt(10000000))
    }
}

func (t *Tool) GetWorkPath() string {
    return t.workPath
}

func (t *Tool) GetDataPath() string {
    if t.netType == 0 {
        return filepath.Join(t.depPath, "data", "mainnet")
    } else if t.netType == 1 {
        return filepath.Join(t.depPath, "data", "testnet")
    } else {
        return filepath.Join(t.depPath, "data", "devnet")
    }
}

func (t *Tool) GetContractPath() string {
    if t.netType == 0 {
        return filepath.Join(t.depPath, "SAFE4-system-contract")
    } else if t.netType == 1 {
        return filepath.Join(t.depPath, "SAFE4-system-contract-testnet")
    } else {
        return filepath.Join(t.depPath, "SAFE4-system-contract-devnet")
    }
}

func (t *Tool) GetSolcPath() string {
    return filepath.Join(t.depPath, "solc.exe")
}

func (t *Tool) GetGenesisPath() string {
    if t.netType == 0 {
        return filepath.Join(t.workPath, "mainnet", "genesis.json")
    } else if t.netType == 1 {
        return filepath.Join(t.workPath, "testnet", "genesis.json")
    } else {
        return filepath.Join(t.workPath, "devnet", "genesis.json")
    }
}

func (t *Tool) GetSafe3StoragePath() string {
    if t.netType == 0 {
        return filepath.Join(t.workPath, "safe3", "safe3storage_mainnet")
    } else if t.netType == 1 {
        return filepath.Join(t.workPath, "safe3", "safe3storage_testnet")
    } else {
        return filepath.Join(t.workPath, "safe3", "safe3storage_devnet")
    }
}

func (t *Tool) GetABI4GoPath() string {
    if t.netType == 0 {
        return filepath.Join(t.workPath, "mainnet", "genesis_abi.go")
    } else if t.netType == 1 {
        return filepath.Join(t.workPath, "testnet", "genesis_abi.go")
    } else {
        return filepath.Join(t.workPath, "devnet", "genesis_abi.go")
    }
}

func (t *Tool) GetABI4JsPath() string {
    if t.netType == 0 {
        return filepath.Join(t.workPath, "mainnet", "genesis_abi.ts")
    } else if t.netType == 1 {
        return filepath.Join(t.workPath, "testnet", "genesis_abi.ts")
    } else {
        return filepath.Join(t.workPath, "devnet", "genesis_abi.ts")
    }
}

func (t *Tool) GetGenesisAlloc() *GenesisAlloc {
    return &t.genesis.Alloc
}

func (t *Tool) GenerateBase() {
    t.genesis.Config = &params.ChainConfig{
        ChainID:             big.NewInt(6666666),
        HomesteadBlock:      big.NewInt(0),
        EIP150Block:         big.NewInt(0),
        EIP150Hash:          common.Hash{},
        EIP155Block:         big.NewInt(0),
        EIP158Block:         big.NewInt(0),
        ByzantiumBlock:      big.NewInt(0),
        ConstantinopleBlock: big.NewInt(0),
        PetersburgBlock:     big.NewInt(0),
        IstanbulBlock:       big.NewInt(0),
        Spos:                &params.SposConfig{Epoch: 200},
    }
    t.genesis.Nonce = "0"
    t.genesis.Timestamp = 0x6375F7B9
    t.genesis.ExtraData = "0x0000000000000000000000000000000000000000000000000000000000000000"
    t.genesis.GasLimit = 0xffffffff
    t.genesis.Difficulty = "0x01"
    t.genesis.Mixhash = common.Hash{}
    t.genesis.Coinbase = common.HexToAddress(t.GetOwnerAddress())
    t.genesis.Number = "0"
    t.genesis.GasUsed = "0"
    t.genesis.ParentHash = common.Hash{}
    t.genesis.Alloc = make(map[common.Address]GenesisAccount)
}

func (t *Tool) setBalance(addr common.Address, balance *big.Int) {
    t.genesis.Alloc[addr] = GenesisAccount{
        Balance: balance.String(),
    }
}

func (t *Tool) AllocBalance() {
    // alloc balance to owner
    t.setBalance(common.HexToAddress(t.GetOwnerAddress()), t.GetOwnerBalance())

    // alloc balance to masternodes
    masternodes := t.loadMasterNode()
    for _, masternode := range *masternodes {
        t.setBalance(masternode.Addr, masternode.Amount)
    }

    // alloc balance to supernodese
    supernodes := t.loadSuperNode()
    for _, supernode := range *supernodes {
        t.setBalance(supernode.Addr, supernode.Amount)
    }
}

func (t *Tool) SaveGenesis() {
    os.MkdirAll(filepath.Dir(t.GetGenesisPath()), 0755)

    b, err := json.Marshal(t.genesis)
    if err != nil {
        panic(err)
    }
    vm := js.New()
    strJS := `function print(str){const obj = JSON.parse(str);return JSON.stringify(obj, null, 2);};print('` + string(b) + `');`
    r, err := vm.RunString(strJS)
    if err != nil {
        panic(err)
    }
    v, _ := r.Export().(string)
    ioutil.WriteFile(t.GetGenesisPath(), []byte(v), 0644)
}

func (t *Tool) SaveABI() {
    os.MkdirAll(filepath.Dir(t.GetABI4GoPath()), 0755)
    os.MkdirAll(filepath.Dir(t.GetABI4JsPath()), 0755)

    contractNames := []string{"Property", "AccountManager", "MasterNodeStorage", "MasterNodeLogic", "SuperNodeStorage", "SuperNodeLogic", "SNVote", "MasterNodeState", "SuperNodeState", "Proposal", "SystemReward", "Safe3", "Multicall"}
    var abis []string
    solcPath := t.GetSolcPath()
    contractPath := t.GetContractPath()
    for _, fileName := range contractNames {
        utils.GetABI(solcPath, contractPath, fileName+".sol")
        abiFile := filepath.Join(contractPath, "temp", fileName+".abi")
        content, err := os.ReadFile(abiFile)
        if err != nil {
            panic(err)
        }
        abis = append(abis, string(content))
        os.RemoveAll(contractPath + "temp")
    }
    temp := "package systemcontracts"
    for i, fileName := range contractNames {
        str, _ := json.Marshal(abis[i])
        temp += fmt.Sprintf("\n\nconst %sABI = %s", fileName, str)
    }

    ioutil.WriteFile(t.GetABI4GoPath(), []byte(temp), 0644)

    temp = ""
    for i, fileName := range contractNames {
        str, _ := json.Marshal(abis[i])
        temp += fmt.Sprintf("export const %sABI = %s as const;", fileName, str[1:len(str)-1])
        if i != len(contractNames)-1 {
            temp += "\n\n"
        }
    }
    temp = strings.Replace(temp, "\\", "", -1)
    ioutil.WriteFile(t.GetABI4JsPath(), []byte(temp), 0644)
}

func (t *Tool) loadMasterNode() *[]MasterNodeInfo {
    jsonFile, err := os.Open(filepath.Join(t.GetDataPath(), "MasterNode.info"))
    if err != nil {
        panic(err)
    }
    defer jsonFile.Close()

    jsonData, err := ioutil.ReadAll(jsonFile)
    if err != nil {
        panic(err)
    }

    masternodes := new([]MasterNodeInfo)
    err = json.Unmarshal(jsonData, masternodes)
    if err != nil {
        panic(err)
    }
    return masternodes
}

func (t *Tool) loadSuperNode() *[]SuperNodeInfo {
    jsonFile, err := os.Open(filepath.Join(t.GetDataPath(), "SuperNode.info"))
    if err != nil {
        panic(err)
    }
    defer jsonFile.Close()

    jsonData, err := ioutil.ReadAll(jsonFile)
    if err != nil {
        panic(err)
    }

    supernodes := new([]SuperNodeInfo)
    err = json.Unmarshal(jsonData, supernodes)
    if err != nil {
        panic(err)
    }
    return supernodes
}

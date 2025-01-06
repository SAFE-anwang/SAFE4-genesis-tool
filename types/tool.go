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
    "runtime"
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
        depPath:   filepath.Join(filepath.Dir(ex), "deps"),
        isStorage: utils.IsStorage(),
    }, err
}

func (t *Tool) GetChainID() *big.Int {
    if t.netType == 0 {
        return big.NewInt(6666665)
    } else {
        return big.NewInt(6666666)
    }
}

func (t *Tool) GetOwnerAddress() string {
    if t.netType == 0 {
        return "0xac110c0f70867f77d9d230e377043f52480a0b7d"
    } else {
        return "0xac110c0f70867f77d9d230e377043f52480a0b7d"
    }
}

func (t *Tool) GetOwnerBalance() *big.Int {
    balance, _ := new(big.Int).SetString("100000000000000000000", 10)
    if t.netType == 0 {
        return balance
    } else {
        return big.NewInt(1).Mul(balance, big.NewInt(10000000000))
    }
}

func (t *Tool) GetWorkPath() string {
    return t.workPath
}

func (t *Tool) GetDataPath() string {
    if t.netType == 0 {
        return filepath.Join(t.depPath, "data", "mainnet")
    } else {
        return filepath.Join(t.depPath, "data", "testnet")
    }
}

func (t *Tool) GetSafe3DataPath() string {
    return filepath.Join(t.depPath, "data", "safe3")
}

func (t *Tool) GetContractPath() string {
    return filepath.Join(t.depPath, "SAFE4-system-contract")
}

func (t *Tool) GetSolcPath() string {
    if runtime.GOOS == "windows" {
        return filepath.Join(t.depPath, "solc-bin", "windows", "solc.exe")
    } else if runtime.GOOS == "linux" {
        return filepath.Join(t.depPath, "solc-bin", "linux", "solc")
    } else if runtime.GOOS == "darwin" {
        return filepath.Join(t.depPath, "solc-bin", "macos", "solc")
    } else {
        panic("Unsupported System")
    }
}

func (t *Tool) GetGenesisPath() string {
    if t.netType == 0 {
        return filepath.Join(t.workPath, "output", "mainnet", "genesis.json")
    } else {
        return filepath.Join(t.workPath, "output", "testnet", "genesis.json")
    }
}

func (t *Tool) GetSafe3StoragePath() string {
    return filepath.Join(t.workPath, "output", "safe3", "safe3storage")
}

func (t *Tool) GetABI4GoPath() string {
    if t.netType == 0 {
        return filepath.Join(t.workPath, "output", "mainnet", "contract_abi.go")
    } else {
        return filepath.Join(t.workPath, "output", "testnet", "contract_abi.go")
    }
}

func (t *Tool) GetABI4JsPath() string {
    if t.netType == 0 {
        return filepath.Join(t.workPath, "output", "mainnet", "contract_abi.ts")
    } else {
        return filepath.Join(t.workPath, "output", "testnet", "contract_abi.ts")
    }
}

func (t *Tool) GetABI4SwiftPath() string {
    if t.netType == 0 {
        return filepath.Join(t.workPath, "output", "mainnet", "Safe4ContractInfo.swift")
    } else {
        return filepath.Join(t.workPath, "output", "testnet", "Safe4ContractInfo.swift")
    }
}

func (t *Tool) GetGenesisAlloc() *GenesisAlloc {
    return &t.genesis.Alloc
}

func (t *Tool) GenerateBase() {
    t.genesis.Config = &params.ChainConfig{
        ChainID:             t.GetChainID(),
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
    // alloc balance to supernodes
    sns := t.loadSuperNode()
    for _, sn := range *sns {
        t.setBalance(sn.Addr, big.NewInt(100000000000000000)) // alloc sn 0.1 for upload-node-state
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
    os.MkdirAll(filepath.Dir(t.GetABI4SwiftPath()), 0755)

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
        os.RemoveAll(filepath.Join(contractPath, "temp"))
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

    temp = "public class Safe4ContractABI {"
    for i, fileName := range contractNames {
        if fileName == "MasterNodeState" || fileName == "SuperNodeState" || fileName == "SystemReward" || fileName == "Multicall" {
            continue
        }
        str, _ := json.Marshal(abis[i])
        temp += fmt.Sprintf("\n    public static var %sABI: String = %s", fileName, str)
    }
    temp += "\n}"
    temp += "\n\npublic class Safe4ContractAddress {\n" +
        "    public static var PropertyContractAddr: String = \"0x0000000000000000000000000000000000001000\"\n" +
        "    public static var AccountManagerContractAddr: String = \"0x0000000000000000000000000000000000001010\"\n" +
        "    public static var MasterNodeStorageContractAddr: String = \"0x0000000000000000000000000000000000001020\"\n" +
        "    public static var MasterNodeLogicContractAddr: String = \"0x0000000000000000000000000000000000001025\"\n" +
        "    public static var SuperNodeStorageContractAddr: String = \"0x0000000000000000000000000000000000001030\"\n" +
        "    public static var SuperNodeLogicContractAddr: String = \"0x0000000000000000000000000000000000001035\"\n" +
        "    public static var SNVoteContractAddr: String = \"0x0000000000000000000000000000000000001040\"\n" +
        "    public static var ProposalContractAddr: String = \"0x0000000000000000000000000000000000001070\"\n" +
        "    public static var Safe3ContractAddr: String = \"0x0000000000000000000000000000000000001090\"\n" +
        "}\n"
    ioutil.WriteFile(t.GetABI4SwiftPath(), []byte(temp), 0644)
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

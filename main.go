package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	//js "github.com/dop251/goja"
	"github.com/safe/SAFE4-genesis-tool/common"
	"github.com/safe/SAFE4-genesis-tool/contracts"
	"github.com/safe/SAFE4-genesis-tool/core"
	"github.com/safe/SAFE4-genesis-tool/params"
	"github.com/safe/SAFE4-genesis-tool/utils"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var workPath string
var ownerAddr string
var genesis core.Genesis
var allocAccounts []common.Address

//var mapAllocAccountStorageKeys map[common.Address][]common.Hash

func main() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	workPath = filepath.Dir(ex) + string(filepath.Separator)
	ownerAddr = utils.GetOwnerAddr()
	autoGenerate()

	autoABI()
	autoABI4JS()
}

func autoGenerate() {
	fmt.Println(time.Now())
	generateBase()
	generateAlloc()
	contracts.NewProxyAdminStorage(workPath, ownerAddr).Generate(&genesis.Alloc)
	contracts.NewPropertyStorage(workPath, ownerAddr).Generate(&genesis.Alloc)
	contracts.NewAccountManagerStorage(workPath, ownerAddr).Generate(&genesis.Alloc)
	contracts.NewMasterNodeStorageStorage(workPath, ownerAddr).Generate(&genesis.Alloc)
	contracts.NewMasterNodeLogicStorage(workPath, ownerAddr).Generate(&genesis.Alloc)
	contracts.NewSuperNodeStorageStorage(workPath, ownerAddr).Generate(&genesis.Alloc)
	contracts.NewSuperNodeLogicStorage(workPath, ownerAddr).Generate(&genesis.Alloc)
	contracts.NewSNVoteStorage(workPath, ownerAddr).Generate(&genesis.Alloc)
	contracts.NewMasterNodeStateStorage(workPath, ownerAddr).Generate(&genesis.Alloc)
	contracts.NewSuperNodeStateStorage(workPath, ownerAddr).Generate(&genesis.Alloc)
	contracts.NewProposalStorage(workPath, ownerAddr).Generate(&genesis.Alloc)
	contracts.NewSystemRewardStorage(workPath, ownerAddr).Generate(&genesis.Alloc)
	contracts.NewSafe3Storage(workPath, ownerAddr).Generate(&genesis.Alloc)
	contracts.NewMulticallStorage(workPath, ownerAddr).Generate(&genesis.Alloc)
	fmt.Println(time.Now())

	b, _ := json.Marshal(genesis)

	//vm := js.New()
	//strJS := `function print(str){const obj = JSON.parse(str);return JSON.stringify(obj, null, 2);};print('` + string(b) + `');`
	//r, err := vm.RunString(strJS)
	//if err != nil {
	//	panic(err)
	//}
	//v, _ := r.Export().(string)
	//ioutil.WriteFile(workPath+utils.GetGenesisFile(), []byte(v), 0644)
	//fmt.Println(time.Now())

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(b); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	ioutil.WriteFile(workPath + utils.GetZipFile(), buf.Bytes(), 0644)
	fmt.Println(time.Now())
}

func generateBase() {
	genesis.Config = &params.ChainConfig{
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
	genesis.Nonce = "0"
	genesis.Timestamp = 0x6375F7B9
	genesis.ExtraData = "0x0000000000000000000000000000000000000000000000000000000000000000"
	genesis.GasLimit = 0xffffffff
	genesis.Difficulty = "2"
	genesis.Mixhash = common.Hash{}
	genesis.Coinbase = common.HexToAddress(ownerAddr)
	genesis.Number = "0"
	genesis.GasUsed = "0"
	genesis.ParentHash = common.Hash{}
	genesis.Alloc = make(map[common.Address]core.GenesisAccount)
}

func allocBalance(addr common.Address, balance *big.Int) {
	account := core.GenesisAccount{
		Balance: balance.String(),
	}
	genesis.Alloc[addr] = account
}

func generateAlloc() {
	// alloc balance to owner
	balance, _ := new(big.Int).SetString("100000000000000000000000", 10)
	allocBalance(common.HexToAddress(ownerAddr), balance)

	// alloc balance to masternodes
	masternodes := contracts.NewAccountManagerStorage(workPath, ownerAddr).LoadMasterNode()
	for _, masternode := range *masternodes {
		allocBalance(masternode.Addr, masternode.Amount)
	}

	// alloc balance to supernodese
	supernodes := contracts.NewAccountManagerStorage(workPath, ownerAddr).LoadSuperNode()
	for _, supernode := range *supernodes {
		allocBalance(supernode.Addr, supernode.Amount)
	}
}

func autoABI() {
	contractNames := []string{"Property", "AccountManager", "MasterNodeStorage", "MasterNodeLogic", "SuperNodeStorage", "SuperNodeLogic", "SNVote", "MasterNodeState", "SuperNodeState", "Proposal", "SystemReward", "Safe3", "Multicall"}
	var abis []string
	for _, fileName := range contractNames {
		utils.GetABI(workPath, fileName+".sol")
		abiFile := workPath + "temp" + string(filepath.Separator) + fileName + ".abi"
		content, err := os.ReadFile(abiFile)
		if err != nil {
			panic(err)
		}
		abis = append(abis, string(content))
		os.RemoveAll(workPath + "temp")
	}
	var temp string
	temp += "package systemcontracts"
	for i, fileName := range contractNames {
		str, _ := json.Marshal(abis[i])
		temp += fmt.Sprintf("\n\nconst %sABI = %s", fileName, str)
	}
	ioutil.WriteFile(workPath+"contract_abi.go", []byte(temp), 0644)
}

func autoABI4JS() {
	contractNames := []string{"Property", "AccountManager", "MasterNodeStorage", "MasterNodeLogic", "SuperNodeStorage", "SuperNodeLogic", "SNVote", "MasterNodeState", "SuperNodeState", "Proposal", "SystemReward", "Safe3"}
	var abis []string
	for _, fileName := range contractNames {
		utils.GetABI(workPath, fileName+".sol")
		abiFile := workPath + "temp" + string(filepath.Separator) + fileName + ".abi"
		content, err := os.ReadFile(abiFile)
		if err != nil {
			panic(err)
		}
		abis = append(abis, string(content))
		os.RemoveAll(workPath + "temp")
	}
	var temp string
	for i, fileName := range contractNames {
		str, _ := json.Marshal(abis[i])
		temp += fmt.Sprintf("export const %sABI = %s as const;", fileName, str[1:len(str)-1])
		if i != len(contractNames)-1 {
			temp += "\n\n"
		}
	}
	temp = strings.Replace(temp, "\\", "", -1)
	ioutil.WriteFile(workPath+"safe4_abi.ts", []byte(temp), 0644)
}

package main

import (
	js "github.com/dop251/goja"
	"github.com/safe/SAFE4-genesis-tool/common"
	"github.com/safe/SAFE4-genesis-tool/common/hexutil"
	"github.com/safe/SAFE4-genesis-tool/contracts"
	"github.com/safe/SAFE4-genesis-tool/core"
	"github.com/safe/SAFE4-genesis-tool/params"
	"github.com/safe/SAFE4-genesis-tool/utils"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
)

var workPath string
var ownerAddr common.Address
var genesis core.Genesis
var allocAccounts []common.Address
var mapAllocAccountStorageKeys map[common.Address][]common.Hash

func main() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	workPath = filepath.Dir(ex) + string(filepath.Separator)
	ownerAddr = common.HexToAddress("0xac110c0f70867f77d9d230e377043f52480a0b7d")
	autoGenerate()
	genesisJson := utils.ToJson(genesis, allocAccounts, mapAllocAccountStorageKeys)

	vm := js.New()
	strJS := `function print(str){const obj = JSON.parse(str);return JSON.stringify(obj, null, 2);};print('` + genesisJson + `');`
	r, err := vm.RunString(strJS)
	if err != nil {
		panic(err)
	}
	v, _ := r.Export().(string)

	err = ioutil.WriteFile(workPath+"genesis.json", []byte(v), 0644)
	if err != nil {
		panic(err)
	}
}

func autoGenerate() {
	generateBase(&allocAccounts)
	mapAllocAccountStorageKeys = make(map[common.Address][]common.Hash)
	contracts.NewSystemStorage(workPath, ownerAddr).Generate(&genesis, &allocAccounts, &mapAllocAccountStorageKeys)
	contracts.NewPropertyStorage(workPath, ownerAddr).Generate(&genesis, &allocAccounts, &mapAllocAccountStorageKeys)
	contracts.NewAccountManagerStorage(workPath, ownerAddr).Generate(&genesis, &allocAccounts, &mapAllocAccountStorageKeys)
	contracts.NewMasterNodeStorage(workPath, ownerAddr).Generate(&genesis, &allocAccounts, &mapAllocAccountStorageKeys)
	contracts.NewSuperNodeStorage(workPath, ownerAddr).Generate(&genesis, &allocAccounts, &mapAllocAccountStorageKeys)
	contracts.NewSNVoteStorage(workPath, ownerAddr).Generate(&genesis, &allocAccounts, &mapAllocAccountStorageKeys)
	contracts.NewMasterNodeStateStorage(workPath, ownerAddr).Generate(&genesis, &allocAccounts, &mapAllocAccountStorageKeys)
	contracts.NewSuperNodeStateStorage(workPath, ownerAddr).Generate(&genesis, &allocAccounts, &mapAllocAccountStorageKeys)
	contracts.NewProposalStorage(workPath, ownerAddr).Generate(&genesis, &allocAccounts, &mapAllocAccountStorageKeys)
	contracts.NewSystemRewardStorage(workPath, ownerAddr).Generate(&genesis, &allocAccounts, &mapAllocAccountStorageKeys)
	contracts.NewSafe3Storage(workPath, ownerAddr).Generate(&genesis, &allocAccounts, &mapAllocAccountStorageKeys)
	contracts.NewMulticallStorage(workPath, ownerAddr).Generate(&genesis, &allocAccounts, &mapAllocAccountStorageKeys)
}

func generateBase(allocAccounts *[]common.Address) {
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
		Spos:                &params.SposConfig{Period: 30, Epoch: 200},
	}
	genesis.Nonce = 0
	genesis.Timestamp = 0x6375F7B9
	genesis.ExtraData = hexutil.MustDecode("0x0000000000000000000000000000000000000000000000000000000000000000")
	genesis.GasLimit = 0xffffffff
	genesis.Difficulty = big.NewInt(2)
	genesis.Mixhash = common.Hash{}
	genesis.Coinbase = ownerAddr
	genesis.Number = 0
	genesis.GasUsed = 0
	genesis.ParentHash = common.Hash{}

	// alloc balance to owner
	genesis.Alloc = core.GenesisAlloc{}
	balance, _ := new(big.Int).SetString("10000000000000000000000", 10)
	genesis.Alloc[ownerAddr] = core.GenesisAccount{Balance: balance}
	*allocAccounts = append(*allocAccounts, ownerAddr)

	// alloc balance to masternodes
	masternodes := contracts.NewAccountManagerStorage(workPath, ownerAddr).LoadMasterNode()
	for _, masternode := range *masternodes {
		genesis.Alloc[masternode.Addr] = core.GenesisAccount{Balance: masternode.Amount}
		*allocAccounts = append(*allocAccounts, masternode.Addr)
	}

	// alloc balance to supernodese
	supernodes := contracts.NewAccountManagerStorage(workPath, ownerAddr).LoadSuperNode()
	for _, supernode := range *supernodes {
		genesis.Alloc[supernode.Addr] = core.GenesisAccount{Balance: supernode.Amount}
		*allocAccounts = append(*allocAccounts, supernode.Addr)
	}
}

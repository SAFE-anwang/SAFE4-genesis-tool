package utils

import (
	"encoding/json"
	"fmt"
	"github.com/safe/SAFE4-genesis-tool/common"
	"github.com/safe/SAFE4-genesis-tool/common/hexutil"
	"github.com/safe/SAFE4-genesis-tool/common/math"
	"github.com/safe/SAFE4-genesis-tool/core"
	"github.com/safe/SAFE4-genesis-tool/params"
	"strconv"
	"strings"
)

type Genesis struct {
	Config     *params.ChainConfig                         	`json:"config"`
	Nonce      string                         				`json:"nonce"`
	Timestamp  math.HexOrDecimal64                         	`json:"timestamp"`
	ExtraData  hexutil.Bytes                               	`json:"extraData"`
	GasLimit   math.HexOrDecimal64                         	`json:"gasLimit"   gencodec:"required"`
	Difficulty string                       				`json:"difficulty" gencodec:"required"`
	Mixhash    common.Hash                                 	`json:"mixHash"`
	Coinbase   common.Address                              	`json:"coinbase"`
	Alloc      core.GenesisAlloc							`json:"alloc"      gencodec:"required"`
	Number     string                         				`json:"number"`
	GasUsed    string                         				`json:"gasUsed"`
	ParentHash common.Hash                                 	`json:"parentHash"`
	BaseFee    *math.HexOrDecimal256                       	`json:"baseFeePerGas,omitempty"`
}

func ToJson(g core.Genesis, allocAccounts []common.Address, mapAllocAccountStorageKeys map[common.Address][]common.Hash) string {
	var enc Genesis
	enc.Config = g.Config
	enc.Nonce = strconv.FormatUint(g.Nonce, 10)
	enc.Timestamp = math.HexOrDecimal64(g.Timestamp)
	enc.ExtraData = g.ExtraData
	enc.GasLimit = math.HexOrDecimal64(g.GasLimit)
	enc.Difficulty = strconv.FormatUint(g.Difficulty.Uint64(), 10)
	enc.Mixhash = g.Mixhash
	enc.Coinbase = g.Coinbase
	enc.Number = strconv.FormatUint(g.Number, 10)
	enc.GasUsed = strconv.FormatUint(g.GasUsed, 10)
	enc.ParentHash = g.ParentHash
	enc.BaseFee = (*math.HexOrDecimal256)(g.BaseFee)
	b, err := json.Marshal(&enc)
	if err != nil {
		panic(err)
	}

	arr := strings.Split(string(b), "null")
	if len(arr) != 2 {
		panic("contain more null filed")
	}

	var genesisJson string
	genesisJson += arr[0]
	genesisJson += allocToJson(g, allocAccounts, mapAllocAccountStorageKeys)
	genesisJson += arr[1]
	return genesisJson
}

func allocToJson(g core.Genesis, allocAccounts []common.Address, mapAllocAccountStorageKeys map[common.Address][]common.Hash) string {
	var allocJson string
	allocJson += fmt.Sprintf("%s", "{")
	for i, addr := range allocAccounts {
		var allocAccountJson string

		// balance
		alloc := g.Alloc[addr]
		allocAccountJson += fmt.Sprintf("\"balance\": \"%s\"", g.Alloc[addr].Balance.String())

		// code
		if len(alloc.Code) != 0 {
			allocAccountJson += fmt.Sprintf(",\"code\": \"0x%s\"", common.Bytes2Hex(alloc.Code))
		}

		// storage
		if alloc.Storage != nil {
			var tempStorageJson string
			storageKeys := mapAllocAccountStorageKeys[addr]
			for k, key := range storageKeys {
				if k != len(storageKeys) - 1 {
					tempStorageJson += fmt.Sprintf("\"%s\": \"%s\",", key.Hex(), alloc.Storage[key].Hex())
				} else {
					tempStorageJson += fmt.Sprintf("\"%s\": \"%s\"", key.Hex(), alloc.Storage[key].Hex())
				}
			}
			allocAccountJson += fmt.Sprintf(",\"storage\": {%s}", tempStorageJson)
		}

		// alloc.account
		if i != len(allocAccounts) - 1 {
			allocJson += fmt.Sprintf("\"%s\": {%s},", addr.Hex(), allocAccountJson)
		} else {
			allocJson += fmt.Sprintf("\"%s\": {%s}", addr.Hex(), allocAccountJson)
		}
	}
	allocJson += fmt.Sprintf("%s", "}")
	return allocJson
}

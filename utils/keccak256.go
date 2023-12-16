package utils

import (
	"github.com/safe/SAFE4-genesis-tool/accounts/abi"
	"github.com/safe/SAFE4-genesis-tool/common"
	"github.com/safe/SAFE4-genesis-tool/crypto"
	"math/big"
)

func Keccak256_uint(slot int64) []byte {
	data := common.BigToHash(big.NewInt(slot))
	return crypto.Keccak256(data.Bytes())
}

func Keccak256_uint_string(slot int64, key string) []byte {
	data1 := []byte(key)
	data2 := common.BigToHash(big.NewInt(slot)).Bytes()
	data := append(data1, data2...)
	return crypto.Keccak256(data)
}

func Keccak256_uint_address(slot int64, key common.Address) []byte {
	uint256Ty, _ := abi.NewType("uint256", "", nil)
	addressTy, _ := abi.NewType("address", "", nil)
	args := abi.Arguments{
		{Type: addressTy},
		{Type: uint256Ty},
	}
	data1 := common.HexToAddress(key.Hex())
	data2 := big.NewInt(slot)
	packed, err := args.Pack(data1, data2)
	if err != nil {
		panic(err)
	}
	return crypto.Keccak256(packed)
}

func Keccak256_uint_uint(slot int64, key int64) []byte {
	data1 := common.BigToHash(big.NewInt(key)).Bytes()
	data2 := common.BigToHash(big.NewInt(slot)).Bytes()
	data := append(data1, data2...)
	return crypto.Keccak256(data)
}

func Keccak256_bytes32(slot string) []byte {
	data, _ := common.ParseHexOrString(slot)
	return crypto.Keccak256(data)
}

func Keccak256_uint_bytes(slot int64, key []byte) []byte {
	data2 := common.BigToHash(big.NewInt(slot)).Bytes()
	data := append(key, data2...)
	return crypto.Keccak256(data)
}
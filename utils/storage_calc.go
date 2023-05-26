package utils

import (
	"github.com/safe/SAFE4-genesis-tool/common"
	"math/big"
)

func GetStorage4String(startKey *big.Int, data string) ([]common.Hash, []common.Hash) {
	var storageKeys, storageValues []common.Hash

	size := len(data)
	if size < 32 {
		// content + size
		storageKey := common.BigToHash(startKey)
		hex := common.Bytes2Hex([]byte(data))
		for k := 0; k < 31 - size; k++ {
			hex += "00"
		}
		hex += common.Bytes2Hex([]byte{byte(size * 2)})
		storageValue := common.HexToHash(hex)
		storageKeys = append(storageKeys, storageKey)
		storageValues = append(storageValues, storageValue)
	} else {
		// size
		storageKey := common.BigToHash(startKey)
		storageValue := common.BigToHash(big.NewInt(int64(size * 2 + 1)))
		storageKeys = append(storageKeys, storageKey)
		storageValues = append(storageValues, storageValue)

		// content
		subKey := big.NewInt(0).SetBytes(Keccak256_bytes32(common.BigToHash(startKey).Hex()))
		descBytes := []byte(data)
		for i := 0; i < size / 32 + 1; i++ {
			tempKey := big.NewInt(0).Add(subKey, big.NewInt(int64(i)))
			start := 32 * i
			end := start + 32
			if end > size {
				end = size
			}
			hex := common.Bytes2Hex(descBytes[start:end])
			for k := 0; k < 32 + start - end; k++ {
				hex += "00"
			}
			subStorageKey := common.BigToHash(tempKey)
			subStorageValue := common.HexToHash(hex)
			storageKeys = append(storageKeys, subStorageKey)
			storageValues = append(storageValues, subStorageValue)
		}
	}

	return storageKeys, storageValues
}

func GetStorage4Int(startKey *big.Int, data *big.Int) (common.Hash, common.Hash) {
	return common.BigToHash(startKey), common.BigToHash(data)
}

func GetStorage4Addr(startKey *big.Int, addr common.Address) (common.Hash, common.Hash) {
	return common.BigToHash(startKey), common.HexToHash(addr.Hex())
}
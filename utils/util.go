package utils

import (
    "math/big"
    "os"
)

func GetDataDir() string {
    for _ ,v := range os.Args {
        if v == "-lmb" {
            return "lmb_data"
        }
    }
    return "data"
}

func GetOwnerAddr() string {
    for _ ,v := range os.Args {
        if v == "-lmb" {
            return "80d8b8f308770ce14252173abb00075cc9082d03"
        }
    }
    return "0xac110c0f70867f77d9d230e377043f52480a0b7d"
}

func GetOwnerBalance() *big.Int {
    for _ ,v := range os.Args {
        if v == "-testnet" || v == "-lmb" {
            balance, _ := new(big.Int).SetString("1000000000000000000000000000000", 10)
            return balance
        }
    }
    balance, _ := new(big.Int).SetString("100000000000000000000000", 10)
    return balance
}

func GetGenesisFile() string {
    for _ ,v := range os.Args {
        if v == "-lmb" {
            return "genesis_lmb.json"
        } else if v == "-testnet" {
            return "genesis_testnet.json"
        }
    }
    return "genesis.json"
}

func IsSaveSafe3Storage() bool {
    for _, v := range os.Args {
        if v == "-safe3" {
            return true
        }
    }
    return false
}

func GetSafe3StorageDir() string {
    for _ ,v := range os.Args {
        if v == "-lmb" {
            return "safe3storage-lmb"
        }
    }
    return "safe3storage"

}
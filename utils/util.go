package utils

import (
    "os"
)

func GetNetType() int {
    for _, v := range os.Args {
        if v == "-devnet" {
            return 2
        }
        if v == "-testnet" {
            return 1
        }
    }
    return 0
}

func IsStorage() bool {
    for _, v := range os.Args {
        if v == "-safe3" {
            return true
        }
    }
    return false
}

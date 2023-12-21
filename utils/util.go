package utils

import "os"

func GetDataDir() string {
    for _ ,v := range os.Args {
        if v == "-test" {
            return "test_data"
        }
    }
    return "data"
}

func GetOwnerAddr() string {
    for _ ,v := range os.Args {
        if v == "-test" {
            return "80d8b8f308770ce14252173abb00075cc9082d03"
        }
    }
    return "0xac110c0f70867f77d9d230e377043f52480a0b7d"
}
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
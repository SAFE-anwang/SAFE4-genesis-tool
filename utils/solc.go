package utils

import (
    "os/exec"
    "path/filepath"
)

func Compile(solcPath string, contractPath string, fileName string) {
    dstPath := filepath.Join(contractPath, "temp")
    openzeppelin_upgradeable_alias := "openzeppelin-contracts-upgradeable/=3rd/OpenZeppelin/openzeppelin-contracts-upgradeable/contracts/"
    openzeppelin_alias := "openzeppelin-contracts/=3rd/OpenZeppelin/openzeppelin-contracts/contracts/"
    cmd := exec.Command(solcPath, "--base-path", contractPath, "--optimize", "--optimize-runs", "200", "--bin-runtime", "-o", dstPath, openzeppelin_upgradeable_alias, openzeppelin_alias, "--overwrite", filepath.Join(contractPath, fileName))
    _, err := cmd.Output()
    if err != nil {
        panic(err)
    }
}

func GetABI(solcPath string, contractPath string, fileName string) {
    dstPath := filepath.Join(contractPath, "temp")
    openzeppelin_upgradeable_alias := "openzeppelin-contracts-upgradeable/=3rd/OpenZeppelin/openzeppelin-contracts-upgradeable/contracts/"
    openzeppelin_alias := "openzeppelin-contracts/=3rd/OpenZeppelin/openzeppelin-contracts/contracts/"
    cmd := exec.Command(solcPath, "--base-path", contractPath, "--optimize", "--optimize-runs", "200", "--abi", "-o", dstPath, openzeppelin_upgradeable_alias, openzeppelin_alias, "--overwrite", filepath.Join(contractPath, fileName))
    _, err := cmd.Output()
    if err != nil {
        panic(err)
    }
}

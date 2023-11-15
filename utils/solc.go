package utils

import (
	"os/exec"
	"path/filepath"
)

func Compile(workPath string, fileName string) {
	solcPath := workPath + "solc.exe"
	filePath := workPath + "SAFE4-system-contract" + string(filepath.Separator)
	basePath := filePath
	dstPath := workPath + "temp"
	openzeppelin_upgradeable_alias := "openzeppelin-contracts-upgradeable/=3rd/OpenZeppelin/openzeppelin-contracts-upgradeable/contracts/"
	openzeppelin_alias := "openzeppelin-contracts/=3rd/OpenZeppelin/openzeppelin-contracts/contracts/"
	cmd := exec.Command(solcPath, "--base-path", basePath, "--optimize-runs", "200", "--bin-runtime", "-o", dstPath, openzeppelin_upgradeable_alias, openzeppelin_alias, "--overwrite", filePath + fileName)
	_, err := cmd.Output()
	if err != nil {
		panic(err)
	}
}

func GetABI(workPath string, fileName string) {
	solcPath := workPath + "solc.exe"
	filePath := workPath + "SAFE4-system-contract" + string(filepath.Separator)
	basePath := filePath
	dstPath := workPath + "temp"
	openzeppelin_upgradeable_alias := "openzeppelin-contracts-upgradeable/=3rd/OpenZeppelin/openzeppelin-contracts-upgradeable/contracts/"
	openzeppelin_alias := "openzeppelin-contracts/=3rd/OpenZeppelin/openzeppelin-contracts/contracts/"
	cmd := exec.Command(solcPath, "--base-path", basePath, "--optimize-runs", "200", "--abi", "-o", dstPath, openzeppelin_upgradeable_alias, openzeppelin_alias, "--overwrite", filePath + fileName)
	_, err := cmd.Output()
	if err != nil {
		panic(err)
	}
}

package utils

import (
	"os/exec"
	"path/filepath"
)

func Compile(workPath string, fileName string) {
	solcPath := workPath + "solc.exe"
	filePath := workPath + "SAFE4-system-contract" + string(filepath.Separator)
	basePath := filePath
	includePath := filePath + ".deps" + string(filepath.Separator) + "npm"
	dstPath := workPath + "temp"
	//cmd := exec.Command(solcPath, "--base-path", basePath, "--include-path", includePath, "--optimize-runs", "200", "--bin-runtime", "--abi", "--storage-layout", "-o", dstPath, "--overwrite", filePath + fileName)
	cmd := exec.Command(solcPath, "--base-path", basePath, "--include-path", includePath, "--optimize-runs", "200", "--bin-runtime", "-o", dstPath, "--overwrite", filePath + fileName)
	_, err := cmd.Output()
	if err != nil {
		panic(err)
	}
}

func GetABI(workPath string, fileName string) {
	solcPath := workPath + "solc.exe"
	filePath := workPath + "SAFE4-system-contract" + string(filepath.Separator)
	basePath := filePath
	includePath := filePath + ".deps" + string(filepath.Separator) + "npm"
	dstPath := workPath + "temp"
	//cmd := exec.Command(solcPath, "--base-path", basePath, "--include-path", includePath, "--optimize-runs", "200", "--bin-runtime", "--abi", "--storage-layout", "-o", dstPath, "--overwrite", filePath + fileName)
	cmd := exec.Command(solcPath, "--base-path", basePath, "--include-path", includePath, "--optimize-runs", "200", "--abi", "-o", dstPath, "--overwrite", filePath + fileName)
	_, err := cmd.Output()
	if err != nil {
		panic(err)
	}
}

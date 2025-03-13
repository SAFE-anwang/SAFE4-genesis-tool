package main

import (
    "github.com/safe/SAFE4-genesis-tool/contracts"
    "github.com/safe/SAFE4-genesis-tool/types"
)

func main() {
    tool, err := types.NewTool()
    if err != nil {
        panic(err)
    }

    tool.GenerateBase()
    tool.AllocBalance()

    contracts.NewProxyAdminStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewPropertyStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewAccountManagerStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewMasterNodeStorageStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewMasterNodeLogicStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewSuperNodeStorageStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewSuperNodeLogicStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewSNVoteStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewMasterNodeStateStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewSuperNodeStateStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewProposalStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewSystemRewardStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewSafe3Storage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewMulticallStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewWSafeStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewMultiSigStorage(tool).Generate(tool.GetGenesisAlloc())
    contracts.NewTimeLockStorage(tool).Generate(tool.GetGenesisAlloc())

    tool.SaveGenesis()
    tool.SaveABI()
}

# SAFE4-genesis-tool
Generator for genesis.json, need depend on SAFE4-system-contract

1. Pull latest commits from SAFE4-system-contract
```
git submodule foreach git fetch
git submodule foreach git pull
```

2. Compile

`go build .`

3. Run

`./SAFE4-genesis-tool [params]`

params:
  
  -testnet: generate genesis.json & ABI files for testnet. Mainnet don't need this parameter.

  -safe3: update safe3storage. If deps/safe3/* is modified, please add this parameter.

4. Output

All run results is saved in output directory.

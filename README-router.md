# How to deploy router swap

## 0. compile

```shell
make all
```
run the above command, it will generate `./build/bin/swaprouter` binary.

## 1. deploy `AnyswapRouter`

deploy a `AnyswapRouter` contract for each supported blockchain

## 2. deploy `AnyswapERC20`

deploy a `AnyswapERC20` contract for each token on each blockchain

## 3. deploy `RouterConfig`

deploy a `RouterConfig` contract to store router bridge configs

## 4. set router config on chain

call `RouterConfig` contract to set configs on blcokchain.

The following is the most used functions, please ref. the abi for more info.

### 4.1 set chain config

call the following contract function:
```solidity
setChainConfig(uint256 chainID, bytes data)
```

data can be generate by the following method:

```shell
./build/bin/swaprouter config genChainConfigData --c.BlockChain fsn --c.ChainID 46688 --c.Confirmations 3 --c.RouterContract 0xc20b5e92e1ce63af6fe537491f75c19016ea5fb4 --c.InitialHeight 3000000 --c.PlusGasPricePercentage 10 --c.MaxGasPriceFluctPercent 20

Output:

chain config struct is {
  "BlockChain": "fsn",
  "ChainID": "46688",
  "Confirmations": 3,
  "RouterContract": "0xc20b5e92e1ce63af6fe537491f75c19016ea5fb4",
  "InitialHeight": 3000000,
  "WaitTimeToReplace": 900,
  "MaxReplaceCount": 20,
  "SwapDeadlineOffset": 36000,
  "PlusGasPricePercentage": 10,
  "MaxGasPriceFluctPercent": 20,
  "DefaultGasLimit": 90000
}
chain config data is 0x7b22426c6f636b436861696e223a2266736e222c22436861696e4944223a223436363838222c22436f6e6669726d6174696f6e73223a332c22526f75746572436f6e7472616374223a22307863323062356539326531636536336166366665353337343931663735633139303136656135666234222c22496e697469616c486569676874223a333030303030302c225761697454696d65546f5265706c616365223a3930302c224d61785265706c616365436f756e74223a32302c2253776170446561646c696e654f6666736574223a33363030302c22506c7573476173507269636550657263656e74616765223a31302c224d61784761735072696365466c75637450657263656e74223a32302c2244656661756c744761734c696d6974223a39303030307d
```

### 4.2 set token config

call the following contract function:
```solidity
setTokenConfig(string tokenID, uint256 chainID, address tokenAddr, bytes data)
```

data can be generate by the following method:

```shell
./build/bin/swaprouter config genTokenConfigData --c.ID any --c.Decimals 18 --c.ContractAddress 0xc20b5e92e1ce63af6fe537491f75c19016ea5fb4 --c.ContractVersion 4 --c.MaximumSwap 1000000 --c.MinimumSwap 100 --c.BigValueThreshold 100000 --c.SwapFeeRate 0.001 --c.MaximumSwapFee 10 --c.MinimumSwapFee 1

Output:

token config struct is {
  "ID": "any",
  "Decimals": 18,
  "ContractAddress": "0xc20b5e92e1ce63af6fe537491f75c19016ea5fb4",
  "ContractVersion": 4,
  "MaximumSwap": 1000000,
  "MinimumSwap": 100,
  "BigValueThreshold": 100000,
  "SwapFeeRate": 0.001,
  "MaximumSwapFee": 10,
  "MinimumSwapFee": 1
}
token config data is 0x7b224944223a22616e79222c22446563696d616c73223a31382c22436f6e747261637441646472657373223a22307863323062356539326531636536336166366665353337343931663735633139303136656135666234222c22436f6e747261637456657273696f6e223a342c224d6178696d756d53776170223a313030303030302c224d696e696d756d53776170223a3130302c2242696756616c75655468726573686f6c64223a3130303030302c225377617046656552617465223a302e3030312c224d6178696d756d53776170466565223a31302c224d696e696d756d53776170466565223a317d
```

### 4.3 set multichain tokens

call the following contract function:
```solidity
addMultichainToken(string tokenID, uint256 chainID, address token)
addMultichainTokens(string tokenID, uint256[] chainIDs, address[] tokens)
```

### 4.4 set mpc address's public key

call the following contract function:
```solidity
setMPCPubkey(address addr, bytes data)
```

data is the hex string of mpc address's public key


## 5. add local config file

please ref. [config-routerswap-example.toml](https://github.com/anyswap/CrossChain-Bridge/blob/router/params/config-routerswap-example.toml)

## 6. run swaprouter

```shell
# for server run (add '--runserver' option)
setsid ./build/bin/swaprouter --config config.toml --log logs/routerswap.log --runserver

# for oracle run
setsid ./build/bin/swaprouter --config config.toml --log logs/routerswap.log
```

## 7. sub commands

get all sub command list and help info, run

```shell
./build/bin/swaprouter -h
```

sub commands:

`admin` is admin tool

`config` is tool for generate chain and token config data

`scanswap` is tool for scan and post swap register

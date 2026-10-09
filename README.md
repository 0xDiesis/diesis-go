<div align="center">

<img src="https://github.com/0xDiesis/diesis/raw/main/docs/diesis.png" alt="Diesis" width="120">

<pre>
 ___ ___ ___ ___ ___ ___
|   \_ _| __/ __|_ _/ __|
| |) | || _|\__ \| |\__ \
|___/___|___|___/___|___/
 -  -  -  -  -  -  -  -
</pre>

**The Go SDK for Diesis**

Call the Diesis system contracts from Go with their addresses, chain data, and
ABIs already in place. Built on
[go-ethereum](https://github.com/ethereum/go-ethereum), so your existing
`ethclient` and signers still work.

Chain ID `1980` · Token **DS** · Runtime **go-ethereum 1.17** · Go **1.25+**

_Greek δίεσις: the smallest interval in music.<br>Diesis aims for the smallest interval between blocks._

</div>

---

## Why this SDK

Diesis builds trading, fee sponsorship, names, staking, and private transfers
into the chain itself, as system contracts at fixed addresses. This module
gives a Go service or bot those addresses and interfaces, so you bind a
contract in one line instead of copying an address and an ABI by hand.

- `diesis.Contracts` pairs each of the 28 fixed-address public contracts with
  its address and ABI.
- The `contracts` package has, for all 29 public contracts, the ABI, the
  canonical signatures, function selectors, and event topics, and structs for
  parameters, events, errors, and named tuples. It is generated from the
  Solidity sources by abi-typegen.
- Addresses and chain data are generated from the same `canonical.json` as the
  [TypeScript SDK](https://github.com/0xDiesis/diesis-js) and the
  [Python SDK](https://github.com/0xDiesis/diesis-py).
- Binding uses go-ethereum's `accounts/abi/bind/v2` package. The older
  `accounts/abi/bind` package is deprecated in current go-ethereum.

## Install

The module is not tagged yet. Fetch it from GitHub with authenticated Git
access, and pin a commit so builds stay reproducible:

```bash
export GOPRIVATE=github.com/0xDiesis/*
go get github.com/0xDiesis/diesis-go@<commit>
```

The module path is `github.com/0xDiesis/diesis-go`. You import it as `diesis`,
and the generated bindings as `github.com/0xDiesis/diesis-go/contracts`.

## Connect

**In short.** You dial a Diesis node, then bind the contract you want at its
fixed address.

**Details.** `diesis.Mainnet` and `diesis.Testnet` carry the chain ID, name,
native currency, RPC URL, and explorer URL. `Contract.Bind` parses the
contract's ABI and returns a `*bind.BoundContract` at its address. It takes
any `bind.ContractBackend`, such as an `*ethclient.Client` or a simulated
backend in tests.

```go
client, err := ethclient.Dial(diesis.Mainnet.RPCURL)
if err != nil {
	log.Fatal(err)
}
defer client.Close()

staking, err := diesis.Contracts.Staking.Bind(client)
if err != nil {
	log.Fatal(err)
}
fmt.Println(staking.Address().Hex())
```

| Network        | Chain ID | Export           |
| -------------- | -------- | ---------------- |
| Diesis Mainnet | 1980     | `diesis.Mainnet` |
| Diesis Testnet | 19803    | `diesis.Testnet` |

## Read a value

**In short.** A view call asks a node for a value without sending a
transaction or paying gas.

**Details.** `BoundContract.Call` packs the arguments, runs `eth_call`, and
unpacks the results into a slice. The generated `<Contract><Method>Params`
structs name each argument and give its Go type, here `TokenId *big.Int`.

```go
staking, err := diesis.Contracts.Staking.Bind(client)
if err != nil {
	log.Fatal(err)
}

args := contracts.DiesisStakingUnclaimedRewardsParams{TokenId: big.NewInt(42)}
var out []any
opts := &bind.CallOpts{Context: context.Background()}
if err := staking.Call(opts, &out, "unclaimedRewards", args.TokenId); err != nil {
	log.Fatal(err)
}
rewards := out[0].(*big.Int)
```

## Send a transaction

**In short.** You sign a transaction with your key, and the node puts it on
chain. Here you stake 100 DS with validator 1.

**Details.** `bind.NewKeyedTransactor` signs locally for the chain ID you
pass. `Chain.ChainID` returns it as a `*big.Int`. `stake` is payable, so the
amount goes in `opts.Value`. Gas, fees, and the nonce are filled in from the
node unless you set them on `opts`.

```go
key, err := crypto.HexToECDSA("<hex private key>")
if err != nil {
	log.Fatal(err)
}

opts := bind.NewKeyedTransactor(key, diesis.Mainnet.ChainID())
opts.Value = new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)) // 100 DS

args := contracts.DiesisStakingStakeParams{ToValidatorId: big.NewInt(1)}
tx, err := staking.Transact(opts, "stake", args.ToValidatorId)
if err != nil {
	log.Fatal(err)
}
receipt, err := bind.WaitMined(context.Background(), client, tx.Hash())
```

## Addresses and contract table

**In short.** Diesis system contracts live at fixed addresses that start with
`0xD1E515`. The SDK has all of them as `common.Address` values, plus a table
that pairs each public contract with its interface.

**Details.** Every `canonical.json` address is a package variable, such as
`diesis.DiesisSpotBook` for `DIESIS_SPOT_BOOK`. `diesis.Addresses` maps the
canonical names to the same values. `diesis.Contracts` has one field per
fixed-address contract, named after the TypeScript SDK's `diesisContracts`
keys (`Staking`, `SpotBook`, `ERC20Factory`, and so on). Each entry has `Key`,
`Name`, `Address`, and `ABI`. `diesis.Contracts.All()` lists them in order.

`IValidatorShare` has no fixed address, so it is not in the table. Bind it by
name with `diesis.Bind`. `diesis.ABIs` and `diesis.ParseABI` cover all 29
public contracts by name.

```go
for _, c := range diesis.Contracts.All() {
	fmt.Println(c.Key, c.Name, c.Address.Hex())
}

share, err := diesis.Bind("IValidatorShare", shareAddress, client)
```

## Selectors, topics, and generated types

**In short.** The `contracts` package names every function, event, and error
of each public contract, so you don't hash signatures or hand-write structs.

**Details.** For each contract `<C>` the package has:

- `<C>ABI`, the JSON ABI.
- `<C><Name>Signature`, the canonical signature, such as
  `DiesisStakingUnclaimedRewardsSignature` for `"unclaimedRewards(uint256)"`.
  Events and errors end in `EventSignature` and `ErrorSignature`.
- `<C><Name>Selector`, the 4-byte function or error selector as a `[4]byte`,
  and `<C><Name>EventTopic`, the event's topic 0 as a `common.Hash`.
  Anonymous events have no topic.
- `<C><Name>Params`, `<C><Name>Event`, and `<C><Name>Error` structs, with
  fields named and typed the way go-ethereum packs and unpacks them.
- One struct per Solidity struct, named after the contract and the struct,
  such as `DiesisStakingIDiesisEpochAuthorityRoleKeyBundle`. Pass these as
  tuple arguments.

Overloaded functions get a numeric suffix, the same one go-ethereum uses in
`abi.ABI.Methods`: `safeTransferFrom(address,address,uint256,bytes)` is
`DiesisStakingSafeTransferFrom0Signature`.

Filter logs by the generated topic, then decode each one into the event
struct with `BoundContract.UnpackLog`.

```go
query := ethereum.FilterQuery{
	FromBlock: big.NewInt(0),
	Addresses: []common.Address{diesis.Contracts.Staking.Address},
	Topics:    [][]common.Hash{{contracts.DiesisStakingStakedEventTopic}},
}
logs, err := client.FilterLogs(context.Background(), query)
if err != nil {
	log.Fatal(err)
}
for _, entry := range logs {
	var ev contracts.DiesisStakingStakedEvent
	if err := staking.UnpackLog(&ev, "Staked", entry); err != nil {
		log.Fatal(err)
	}
	fmt.Println(ev.Delegator.Hex(), ev.ToValidatorId, ev.Amount)
}
```

The ABI constants also work with plain go-ethereum, for packing calldata
yourself. The first four bytes of the calldata are the generated selector.

```go
parsed, err := abi.JSON(strings.NewReader(contracts.DiesisStakingABI))
if err != nil {
	log.Fatal(err)
}
data, err := parsed.Pack("unclaimedRewards", big.NewInt(42))
if err != nil {
	log.Fatal(err)
}
fmt.Println([4]byte(data[:4]) == contracts.DiesisStakingUnclaimedRewardsSelector) // true
```

All of these examples compile as `Example` functions in `example_test.go`.

## Regenerate

**In short.** The bindings, addresses, chains, and contract table are
generated. Don't edit them by hand. Rerun the script instead.

**Details.** `scripts/generate.py` needs only Python 3, `gofmt`, abi-typegen
0.8.0 or newer, and a built contracts checkout. It runs
`abi-typegen generate --target go --package contracts` for the public
contracts into a temporary directory, formats the result with `gofmt`, and
writes it to `contracts/`. It then copies
`../diesis-js/canonical.json` into the repo and generates `addresses.go`,
`chains.go`, and `contract_table.go` from it. `--check` writes nothing and
fails if any of those files, or the vendored `canonical.json`, would change.

```bash
python3 scripts/generate.py          # regenerate everything
python3 scripts/generate.py --check  # fail if anything is stale
```

The contracts checkout defaults to `../../diesis-core/diesis/contracts`. Set
`DIESIS_CONTRACTS_DIR` to use another one, and `ABI_TYPEGEN` to use a
different `abi-typegen` binary than `<contracts>/node_modules/.bin/abi-typegen`.

## Develop

```bash
go vet ./...
gofmt -l .     # prints nothing when formatted
go test ./...
python3 scripts/generate.py --check
```

## License

MIT

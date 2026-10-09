package diesis_test

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	diesis "github.com/0xDiesis/diesis-go"
	"github.com/0xDiesis/diesis-go/contracts"
)

// Connect to a node and bind the staking contract at its fixed address.
func Example_connect() {
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
}

// Read a value with a view call.
func Example_read() {
	client, err := ethclient.Dial(diesis.Mainnet.RPCURL)
	if err != nil {
		log.Fatal(err)
	}
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
	fmt.Println(rewards)
}

// Sign and send a transaction. stake is payable, so the amount goes in Value.
func Example_sendTransaction() {
	client, err := ethclient.Dial(diesis.Mainnet.RPCURL)
	if err != nil {
		log.Fatal(err)
	}
	key, err := crypto.HexToECDSA("<hex private key>")
	if err != nil {
		log.Fatal(err)
	}
	staking, err := diesis.Contracts.Staking.Bind(client)
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
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(receipt.Status)
}

// Walk the contract table, and bind a contract that has no fixed address.
func Example_contractTable() {
	for _, c := range diesis.Contracts.All()[:3] {
		fmt.Println(c.Key, c.Name, c.Address.Hex())
	}
	fmt.Println(diesis.Addresses["DIESIS_SPOT_BOOK"] == diesis.DiesisSpotBook)

	client, err := ethclient.Dial(diesis.Testnet.RPCURL)
	if err != nil {
		log.Fatal(err)
	}
	shareAddress := common.HexToAddress("0x0000000000000000000000000000000000001234")
	share, err := diesis.Bind("IValidatorShare", shareAddress, client)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(share.Address() == shareAddress)
	// Output:
	// markets IDiesisMarkets 0xd1E515000000000000000000000000000000B00c
	// spotBook IDiesisSpotBook 0xD1E515000000000000000000000000000000590d
	// perpsBook IDiesisPerpsBook 0xD1e5150000000000000000000000000000009E29
	// true
	// true
}

// Find Staked events by their generated topic and decode each one into the
// generated event struct.
func Example_decodeLogs() {
	client, err := ethclient.Dial(diesis.Mainnet.RPCURL)
	if err != nil {
		log.Fatal(err)
	}
	staking, err := diesis.Contracts.Staking.Bind(client)
	if err != nil {
		log.Fatal(err)
	}

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
}

// Use a generated ABI constant directly with go-ethereum. The generated
// selector is the first four bytes of the calldata.
func Example_packCalldata() {
	parsed, err := abi.JSON(strings.NewReader(contracts.DiesisStakingABI))
	if err != nil {
		log.Fatal(err)
	}
	data, err := parsed.Pack("unclaimedRewards", big.NewInt(42))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(contracts.DiesisStakingUnclaimedRewardsSignature)
	fmt.Println([4]byte(data[:4]) == contracts.DiesisStakingUnclaimedRewardsSelector)
	// Output:
	// unclaimedRewards(uint256)
	// true
}

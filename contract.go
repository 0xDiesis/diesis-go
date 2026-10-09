package diesis

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	"github.com/ethereum/go-ethereum/common"
)

// Contract is a public system contract at a fixed address.
type Contract struct {
	// Key is the TypeScript SDK's diesisContracts key, such as "staking".
	Key string
	// Name is the Solidity contract or interface name, such as "DiesisStaking".
	Name string
	// Address is the contract's fixed address.
	Address common.Address
	// ABI is the generated ABI JSON from the contracts package.
	ABI string
}

// ParseABI parses the contract's ABI.
func (c Contract) ParseABI() (abi.ABI, error) {
	return parseABI(c.Name, c.ABI)
}

// Bind returns a bound contract at the contract's fixed address. Pass an
// *ethclient.Client or any other bind.ContractBackend.
func (c Contract) Bind(backend bind.ContractBackend) (*bind.BoundContract, error) {
	return bindABI(c.Name, c.ABI, c.Address, backend)
}

// ParseABI parses the ABI of the named public contract, such as
// "IValidatorShare".
func ParseABI(name string) (abi.ABI, error) {
	raw, ok := ABIs[name]
	if !ok {
		return abi.ABI{}, fmt.Errorf("diesis: unknown contract %q", name)
	}
	return parseABI(name, raw)
}

// Bind returns a bound contract for the named public contract at address.
// Use it for contracts without a fixed address, such as IValidatorShare.
func Bind(name string, address common.Address, backend bind.ContractBackend) (*bind.BoundContract, error) {
	raw, ok := ABIs[name]
	if !ok {
		return nil, fmt.Errorf("diesis: unknown contract %q", name)
	}
	return bindABI(name, raw, address, backend)
}

func parseABI(name, raw string) (abi.ABI, error) {
	parsed, err := abi.JSON(strings.NewReader(raw))
	if err != nil {
		return abi.ABI{}, fmt.Errorf("diesis: parse %s ABI: %w", name, err)
	}
	return parsed, nil
}

func bindABI(name, raw string, address common.Address, backend bind.ContractBackend) (*bind.BoundContract, error) {
	parsed, err := parseABI(name, raw)
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, backend, backend, backend), nil
}

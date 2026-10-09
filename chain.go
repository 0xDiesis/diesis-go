package diesis

import "math/big"

// NativeCurrency describes a chain's native token.
type NativeCurrency struct {
	Name     string
	Symbol   string
	Decimals uint8
}

// Chain describes a Diesis network.
type Chain struct {
	ID             uint64
	Name           string
	NativeCurrency NativeCurrency
	RPCURL         string
	ExplorerURL    string
	Testnet        bool
}

// ChainID returns the chain ID as a *big.Int, the form go-ethereum signers
// and transactors take.
func (c Chain) ChainID() *big.Int {
	return new(big.Int).SetUint64(c.ID)
}

package diesis_test

import (
	"encoding/json"
	"math/big"
	"os"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	diesis "github.com/0xDiesis/diesis-go"
	"github.com/0xDiesis/diesis-go/contracts"
)

type canonicalChain struct {
	ID             uint64 `json:"id"`
	Name           string `json:"name"`
	NativeCurrency struct {
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		Decimals uint8  `json:"decimals"`
	} `json:"nativeCurrency"`
	RPCURL      string `json:"rpcUrl"`
	ExplorerURL string `json:"explorerUrl"`
	Testnet     bool   `json:"testnet"`
}

type canonicalFile struct {
	SchemaVersion int                       `json:"schemaVersion"`
	Chains        map[string]canonicalChain `json:"chains"`
	Addresses     map[string]string         `json:"addresses"`
}

func loadCanonical(t *testing.T) canonicalFile {
	t.Helper()
	raw, err := os.ReadFile("canonical.json")
	if err != nil {
		t.Fatal(err)
	}
	var c canonicalFile
	if err := json.Unmarshal(raw, &c); err != nil {
		t.Fatal(err)
	}
	return c
}

// The fixed-address public contracts, keyed as in the TypeScript SDK's
// diesisContracts, with their contract name and canonical.json address name.
var fixedContracts = []struct{ key, name, address string }{
	{"markets", "IDiesisMarkets", "DIESIS_MARKETS"},
	{"spotBook", "IDiesisSpotBook", "DIESIS_SPOT_BOOK"},
	{"perpsBook", "IDiesisPerpsBook", "DIESIS_PERPS_BOOK"},
	{"margin", "IDiesisMargin", "DIESIS_MARGIN"},
	{"settlement", "IDiesisSettlement", "DIESIS_SETTLEMENT"},
	{"settlementRouter", "DiesisSettlementRouter", "DIESIS_SETTLEMENT_ROUTER"},
	{"conductors", "IDiesisConductors", "DIESIS_CONDUCTORS"},
	{"erc20Factory", "IDiesisErc20Factory", "DIESIS_ERC20_FACTORY"},
	{"perpDeploy", "IDiesisPerpDeploy", "DIESIS_PERP_DEPLOY"},
	{"operatorBond", "IDiesisOperatorBond", "DIESIS_OPERATOR_BOND"},
	{"bundleEscrow", "IDiesisBundleEscrow", "DIESIS_BUNDLE_ESCROW"},
	{"staking", "DiesisStaking", "DIESIS_STAKING"},
	{"position", "IDiesisPosition", "DIESIS_POSITION"},
	{"liquidStakedDS", "ILiquidStakedDS", "LIQUID_STAKED_DS"},
	{"wrappedDS", "IWrappedDS", "WRAPPED_DS"},
	{"patron", "DiesisPatron", "DIESIS_PATRON"},
	{"config", "DiesisConfig", "DIESIS_CONFIG"},
	{"coreVault", "IDiesisCoreVault", "DIESIS_CORE_VAULT"},
	{"issuanceAuction", "IDiesisIssuanceAuction", "DIESIS_ISSUANCE_AUCTION"},
	{"buybackBurn", "IDiesisBuybackBurn", "DIESIS_BUYBACK_BURN"},
	{"shieldedPool", "DiesisShieldedPool", "SHIELDED_POOL"},
	{"privacyPools", "DiesisPrivacyPools", "PRIVACY_POOLS"},
	{"nameRegistry", "IDiesisNameRegistry", "DIESIS_NAME_REGISTRY"},
	{"baseRegistrar", "DiesisBaseRegistrar", "DIESIS_BASE_REGISTRAR"},
	{"publicResolver", "DiesisPublicResolver", "DIESIS_PUBLIC_RESOLVER"},
	{"reverseRegistrar", "IDiesisReverseRegistrar", "DIESIS_REVERSE_REGISTRAR"},
	{"nameVerifier", "IDiesisNameVerifier", "DIESIS_NAME_VERIFIER"},
	{"namePolicy", "IDiesisNamePolicy", "DIESIS_NAME_POLICY"},
}

func TestAddressesMatchCanonical(t *testing.T) {
	c := loadCanonical(t)
	if len(diesis.Addresses) != len(c.Addresses) {
		t.Fatalf("got %d addresses, canonical.json has %d", len(diesis.Addresses), len(c.Addresses))
	}
	for name, want := range c.Addresses {
		got, ok := diesis.Addresses[name]
		if !ok {
			t.Errorf("%s missing", name)
			continue
		}
		if !strings.EqualFold(got.Hex(), want) {
			t.Errorf("%s = %s, want %s", name, got.Hex(), want)
		}
	}
	if diesis.DiesisStaking != common.HexToAddress(c.Addresses["DIESIS_STAKING"]) {
		t.Errorf("DiesisStaking = %s", diesis.DiesisStaking.Hex())
	}
	if diesis.DiesisSpotBook != common.HexToAddress(c.Addresses["DIESIS_SPOT_BOOK"]) {
		t.Errorf("DiesisSpotBook = %s", diesis.DiesisSpotBook.Hex())
	}
}

func TestChains(t *testing.T) {
	c := loadCanonical(t)
	if diesis.SchemaVersion != c.SchemaVersion {
		t.Errorf("SchemaVersion = %d, want %d", diesis.SchemaVersion, c.SchemaVersion)
	}
	if diesis.MainnetChainID != 1980 || diesis.TestnetChainID != 19803 {
		t.Errorf("chain IDs = %d, %d", diesis.MainnetChainID, diesis.TestnetChainID)
	}
	for key, got := range map[string]diesis.Chain{"mainnet": diesis.Mainnet, "testnet": diesis.Testnet} {
		want := c.Chains[key]
		if got.ID != want.ID || got.Name != want.Name || got.RPCURL != want.RPCURL ||
			got.ExplorerURL != want.ExplorerURL || got.Testnet != want.Testnet ||
			got.NativeCurrency.Name != want.NativeCurrency.Name ||
			got.NativeCurrency.Symbol != want.NativeCurrency.Symbol ||
			got.NativeCurrency.Decimals != want.NativeCurrency.Decimals {
			t.Errorf("%s = %+v, want %+v", key, got, want)
		}
	}
	if diesis.Mainnet.Name != "Diesis" || diesis.Testnet.Name != "Diesis Testnet" {
		t.Errorf("names = %q, %q", diesis.Mainnet.Name, diesis.Testnet.Name)
	}
	if diesis.Mainnet.NativeCurrency.Symbol != "DS" || diesis.Mainnet.NativeCurrency.Decimals != 18 {
		t.Errorf("native currency = %+v", diesis.Mainnet.NativeCurrency)
	}
	if diesis.Testnet.ChainID().Uint64() != 19803 {
		t.Errorf("Testnet.ChainID() = %s", diesis.Testnet.ChainID())
	}
	if len(diesis.Chains) != 2 || diesis.Chains[0].ID != 1980 {
		t.Errorf("Chains = %+v", diesis.Chains)
	}
}

func TestContractTable(t *testing.T) {
	c := loadCanonical(t)
	all := diesis.Contracts.All()
	if len(all) != 28 || len(all) != len(fixedContracts) {
		t.Fatalf("table has %d entries, want 28", len(all))
	}
	for i, want := range fixedContracts {
		got := all[i]
		if got.Key != want.key || got.Name != want.name {
			t.Errorf("entry %d = %s/%s, want %s/%s", i, got.Key, got.Name, want.key, want.name)
		}
		if got.Address != common.HexToAddress(c.Addresses[want.address]) {
			t.Errorf("%s address = %s, want %s", want.key, got.Address.Hex(), c.Addresses[want.address])
		}
		if got.ABI != diesis.ABIs[want.name] {
			t.Errorf("%s ABI does not match the generated %sABI", want.key, want.name)
		}
	}
	if diesis.Contracts.Staking.Address != diesis.DiesisStaking || diesis.Contracts.Staking.ABI != contracts.DiesisStakingABI {
		t.Errorf("Contracts.Staking = %+v", diesis.Contracts.Staking.Address)
	}
}

func TestEveryPublicABIParses(t *testing.T) {
	if len(diesis.ABIs) != 29 {
		t.Fatalf("ABIs has %d entries, want 29", len(diesis.ABIs))
	}
	for name, raw := range diesis.ABIs {
		parsed, err := abi.JSON(strings.NewReader(raw))
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		if len(parsed.Methods) == 0 {
			t.Errorf("%s has no methods", name)
		}
	}
	staking, err := diesis.Contracts.Staking.ParseABI()
	if err != nil {
		t.Fatal(err)
	}
	for _, method := range []string{"stake", "unclaimedRewards", "nodeLedger"} {
		if _, ok := staking.Methods[method]; !ok {
			t.Errorf("DiesisStaking ABI has no %s", method)
		}
	}
	if _, err := diesis.ParseABI("IValidatorShare"); err != nil {
		t.Error(err)
	}
	if _, err := diesis.ParseABI("NotAContract"); err == nil {
		t.Error("ParseABI accepted an unknown contract")
	}
}

func TestBind(t *testing.T) {
	bound, err := diesis.Contracts.Staking.Bind(nil)
	if err != nil {
		t.Fatal(err)
	}
	if bound.Address() != diesis.DiesisStaking {
		t.Errorf("bound address = %s", bound.Address().Hex())
	}
	share := common.HexToAddress("0x0000000000000000000000000000000000001234")
	bound, err = diesis.Bind("IValidatorShare", share, nil)
	if err != nil {
		t.Fatal(err)
	}
	if bound.Address() != share {
		t.Errorf("bound address = %s", bound.Address().Hex())
	}
	if _, err := diesis.Bind("NotAContract", share, nil); err == nil {
		t.Error("Bind accepted an unknown contract")
	}
}

func TestSelectorsAndTopics(t *testing.T) {
	staking, err := diesis.Contracts.Staking.ParseABI()
	if err != nil {
		t.Fatal(err)
	}
	methods := []struct {
		name      string
		signature string
		selector  [4]byte
	}{
		{"stake", contracts.DiesisStakingStakeSignature, contracts.DiesisStakingStakeSelector},
		{"unclaimedRewards", contracts.DiesisStakingUnclaimedRewardsSignature, contracts.DiesisStakingUnclaimedRewardsSelector},
		{"registerValidatorRoleKeys", contracts.DiesisStakingRegisterValidatorRoleKeysSignature, contracts.DiesisStakingRegisterValidatorRoleKeysSelector},
		// go-ethereum names the second overload safeTransferFrom0, as abi-typegen does.
		{"safeTransferFrom0", contracts.DiesisStakingSafeTransferFrom0Signature, contracts.DiesisStakingSafeTransferFrom0Selector},
	}
	for _, m := range methods {
		method, ok := staking.Methods[m.name]
		if !ok {
			t.Errorf("DiesisStaking ABI has no %s", m.name)
			continue
		}
		if method.Sig != m.signature {
			t.Errorf("%s signature = %q, go-ethereum has %q", m.name, m.signature, method.Sig)
		}
		if [4]byte(method.ID) != m.selector {
			t.Errorf("%s selector = %x, go-ethereum has %x", m.name, m.selector, method.ID)
		}
	}
	staked := staking.Events["Staked"]
	if staked.Sig != contracts.DiesisStakingStakedEventSignature || staked.ID != contracts.DiesisStakingStakedEventTopic {
		t.Errorf("Staked = %q %s, go-ethereum has %q %s", contracts.DiesisStakingStakedEventSignature,
			contracts.DiesisStakingStakedEventTopic.Hex(), staked.Sig, staked.ID.Hex())
	}
}

func TestBundleEscrowInitialWireSelectors(t *testing.T) {
	escrow, err := diesis.Contracts.BundleEscrow.ParseABI()
	if err != nil {
		t.Fatal(err)
	}
	methods := []struct {
		name      string
		signature string
		selector  [4]byte
		generated [4]byte
	}{
		{"reserveBundle", contracts.IDiesisBundleEscrowReserveBundleSignature, [4]byte{0x79, 0x3b, 0xae, 0xf2}, contracts.IDiesisBundleEscrowReserveBundleSelector},
		{"finalizeBundle", contracts.IDiesisBundleEscrowFinalizeBundleSignature, [4]byte{0x1e, 0x73, 0xb2, 0xe6}, contracts.IDiesisBundleEscrowFinalizeBundleSelector},
		{"cancelBundle", contracts.IDiesisBundleEscrowCancelBundleSignature, [4]byte{0xd3, 0x07, 0xc0, 0xa3}, contracts.IDiesisBundleEscrowCancelBundleSelector},
		{"reclaimExpiredBundle", contracts.IDiesisBundleEscrowReclaimExpiredBundleSignature, [4]byte{0x3d, 0xc7, 0x27, 0xf7}, contracts.IDiesisBundleEscrowReclaimExpiredBundleSelector},
	}
	for _, expected := range methods {
		method, ok := escrow.Methods[expected.name]
		if !ok {
			t.Errorf("bundle escrow ABI has no %s", expected.name)
			continue
		}
		if method.Sig != expected.signature || [4]byte(method.ID) != expected.selector || expected.generated != expected.selector {
			t.Errorf("%s = %s %x, want %s %x", expected.name, method.Sig, method.ID, expected.signature, expected.selector)
		}
	}
	for _, old := range []string{"reserveBundleV2", "finalizeBundleV2", "cancelBundleV2", "reclaimExpiredBundleV2"} {
		if _, exists := escrow.Methods[old]; exists {
			t.Errorf("obsolete bundle method %s remains", old)
		}
	}
}

func TestTupleStructsPack(t *testing.T) {
	staking, err := diesis.Contracts.Staking.ParseABI()
	if err != nil {
		t.Fatal(err)
	}
	args := contracts.DiesisStakingRegisterValidatorRoleKeysParams{
		ValidatorId: big.NewInt(7),
		TargetEpoch: 3,
		Keys: contracts.DiesisStakingIDiesisEpochAuthorityRoleKeyBundle{
			ConsensusAddress: common.HexToAddress("0x01"),
			BlsPublicKey:     []byte{1, 2, 3},
		},
		Proofs: contracts.DiesisStakingIDiesisEpochAuthorityRoleProofBundle{
			VrfX: big.NewInt(1),
			VrfY: big.NewInt(2),
		},
	}
	data, err := staking.Pack("registerValidatorRoleKeys", args.ValidatorId, args.TargetEpoch, args.Keys, args.Proofs)
	if err != nil {
		t.Fatal(err)
	}
	if [4]byte(data[:4]) != contracts.DiesisStakingRegisterValidatorRoleKeysSelector {
		t.Errorf("calldata starts with %x", data[:4])
	}
	decoded, err := staking.Methods["registerValidatorRoleKeys"].Inputs.Unpack(data[4:])
	if err != nil {
		t.Fatal(err)
	}
	var keys contracts.DiesisStakingIDiesisEpochAuthorityRoleKeyBundle
	abi.ConvertType(decoded[2], &keys)
	if keys.ConsensusAddress != args.Keys.ConsensusAddress || string(keys.BlsPublicKey) != string(args.Keys.BlsPublicKey) {
		t.Errorf("keys round trip = %+v", keys)
	}
}

func TestEventStructUnpacksLog(t *testing.T) {
	staking, err := diesis.Contracts.Staking.Bind(nil)
	if err != nil {
		t.Fatal(err)
	}
	delegator := common.HexToAddress("0x00000000000000000000000000000000000000aa")
	entry := types.Log{
		Address: diesis.DiesisStaking,
		Topics: []common.Hash{
			contracts.DiesisStakingStakedEventTopic,
			common.BytesToHash(delegator.Bytes()),
			common.BigToHash(big.NewInt(1)),
			common.BigToHash(big.NewInt(42)),
		},
		Data: common.BigToHash(big.NewInt(1000)).Bytes(),
	}
	var ev contracts.DiesisStakingStakedEvent
	if err := staking.UnpackLog(&ev, "Staked", entry); err != nil {
		t.Fatal(err)
	}
	if ev.Delegator != delegator || ev.ToValidatorId.Int64() != 1 || ev.TokenId.Int64() != 42 || ev.Amount.Int64() != 1000 {
		t.Errorf("Staked = %+v", ev)
	}
}

package ordinals

import (
	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
)

// Utxo represents an unspent transaction output
type Utxo struct {
	TxID         string
	Vout         uint32
	ScriptPubKey string
	Satoshis     uint64
}

// PayToAddress represents a destination for payment
type PayToAddress struct {
	Address  string
	Satoshis uint64
}

// File represents a file for inscription
type File struct {
	Content     []byte
	ContentType string
}

// TokenType represents the type of token
type TokenType string

const (
	// TokenTypeBSV21 represents BSV21 tokens
	TokenTypeBSV21 TokenType = "bsv-21"
)

// TokenInputMode represents how token inputs are consumed
type TokenInputMode string

const (
	// TokenInputModeAll consumes all token inputs
	TokenInputModeAll TokenInputMode = "all"
	// TokenInputModeNeeded consumes only what's needed for the transaction
	TokenInputModeNeeded TokenInputMode = "needed"
)

// TokenSplitConfig represents configuration for splitting token outputs
type TokenSplitConfig struct {
	// Outputs is the number of outputs to split tokens into
	Outputs int
	// Threshold is the minimum amount of tokens per output
	Threshold *float64
	// OmitMetadata determines whether to omit metadata from token change outputs
	OmitMetadata bool
}

// NftUtxo represents an NFT utxo
type NftUtxo struct {
	Utxo
	ContentType  string
	CollectionID string
}

// TokenUtxo represents a token utxo
type TokenUtxo struct {
	Utxo
	TokenID  string
	Protocol TokenType
	Amount   uint64
	Decimals uint8
}

// TokenDistribution represents a token distribution
type TokenDistribution struct {
	Address string
	Tokens  float64
}

// CreateOrdinalsConfig represents configuration for creating ordinals
type CreateOrdinalsConfig struct {
	Utxos        []*Utxo
	Destinations []*struct {
		Address  string
		File     *File
		Metadata map[string][]byte
	}
	PaymentPk     *ec.PrivateKey
	ChangeAddress string
	SatsPerKb     uint64
}

// SendOrdinalsConfig represents configuration for sending ordinals
type SendOrdinalsConfig struct {
	PaymentUtxos []*Utxo
	Ordinals     []*NftUtxo
	PaymentPk    *ec.PrivateKey
	OrdPk        *ec.PrivateKey
	Destinations []*struct {
		Address  string
		File     *File
		Metadata map[string][]byte
	}
	ChangeAddress string
	SatsPerKb     uint64
}

// SendUtxosConfig represents configuration for sending utxos
type SendUtxosConfig struct {
	Utxos         []*Utxo
	PaymentPk     *ec.PrivateKey
	Payments      []*PayToAddress
	ChangeAddress string
	SatsPerKb     uint64
}

// DeployBsv21TokenConfig represents configuration for deploying a BSV21 token
type DeployBsv21TokenConfig struct {
	Symbol              string
	Icon                string
	Utxos               []*Utxo
	InitialDistribution *TokenDistribution
	PaymentPk           *ec.PrivateKey
	DestinationAddress  string
	ChangeAddress       string
	SatsPerKb           uint64
}

// TransferBsv21TokenConfig represents configuration for transferring BSV21 tokens
type TransferBsv21TokenConfig struct {
	Protocol      TokenType
	TokenID       string
	Utxos         []*Utxo
	InputTokens   []*TokenUtxo
	Distributions []*TokenDistribution
	PaymentPk     *ec.PrivateKey
	OrdPk         *ec.PrivateKey
	Burn          bool
	ChangeAddress string
	SatsPerKb     uint64
	// TokenInputMode determines how token inputs are consumed (all or only what's needed)
	TokenInputMode TokenInputMode
	// SplitConfig configures how token change outputs are split
	SplitConfig *TokenSplitConfig
	// Decimals is the number of decimal places for the token
	Decimals uint8
}

// CreateOrdListingsConfig represents configuration for creating ordinal listings
type CreateOrdListingsConfig struct {
	Utxos    []*Utxo
	Listings []*struct {
		PayAddress  string
		Price       uint64
		ListingUtxo *NftUtxo
		OrdAddress  string
	}
	PaymentPk     *ec.PrivateKey
	OrdPk         *ec.PrivateKey
	ChangeAddress string
	SatsPerKb     uint64
}

// PurchaseOrdListingConfig represents configuration for purchasing an ordinal listing
type PurchaseOrdListingConfig struct {
	Utxos         []*Utxo
	PaymentPk     *ec.PrivateKey
	ListingUtxo   *NftUtxo
	OrdAddress    string
	ChangeAddress string
	SatsPerKb     uint64
}

// CancelOrdListingsConfig represents configuration for cancelling ordinal listings
type CancelOrdListingsConfig struct {
	Utxos         []*Utxo
	ListingUtxos  []*NftUtxo
	OrdPk         *ec.PrivateKey
	PaymentPk     *ec.PrivateKey
	ChangeAddress string
	SatsPerKb     uint64
}

// BroadcastResult represents the result of broadcasting a transaction
type BroadcastResult struct {
	Status  string
	TxID    string
	Message string
}

// BroadcastFunc is a function type for broadcasting transactions
type BroadcastFunc func(tx *transaction.Transaction) (*BroadcastResult, error)

// CreateUnlocker creates an unlocker for a private key
func CreateUnlocker(privateKey *ec.PrivateKey) (interface{}, error) {
	// We'll implement this in a utility file
	return nil, nil
}

// Helper function to convert an address string to a script.Address
func AddressFromString(addressStr string) (*script.Address, error) {
	return script.NewAddressFromString(addressStr)
}

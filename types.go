package ordinals

import (
	"github.com/bitcoin-sv/go-templates/template/inscription"
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

// Destination represents a destination for an inscription
type Destination struct {
	// Address is the destination address for the inscription
	Address string
	// Inscription is the optional inscription to include
	Inscription *inscription.Inscription
	// omitMetadata determines whether to omit metadata from this inscription
	omitMetadata bool
}

// OmitMetadata returns whether metadata should be omitted from this inscription
func (d *Destination) OmitMetadata() bool {
	return d.omitMetadata
}

// SetOmitMetadata sets whether metadata should be omitted from this inscription
func (d *Destination) SetOmitMetadata(omit bool) {
	d.omitMetadata = omit
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
	// OmitMetadata determines whether to omit metadata from this distribution's output
	OmitMetadata bool
}

// CreateOrdinalsConfig represents configuration for creating ordinals
type CreateOrdinalsConfig struct {
	Utxos         []*Utxo
	Destinations  []*Destination
	PaymentPk     *ec.PrivateKey
	ChangeAddress string
	SatsPerKb     uint64
	// AdditionalPayments is an optional list of additional payments to make
	AdditionalPayments []*PayToAddress
}

// SendOrdinalsConfig represents configuration for sending ordinals
type SendOrdinalsConfig struct {
	PaymentUtxos  []*Utxo
	Ordinals      []*NftUtxo
	PaymentPk     *ec.PrivateKey
	OrdPk         *ec.PrivateKey
	Destinations  []*Destination
	ChangeAddress string
	SatsPerKb     uint64
	// AdditionalPayments is an optional list of additional payments to make
	AdditionalPayments []*PayToAddress
	// EnforceUniformSend ensures that the number of destinations matches the number of ordinals
	EnforceUniformSend bool
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

// CreateOrdTokenListingsConfig represents configuration for creating token listings
type CreateOrdTokenListingsConfig struct {
	// Utxos are the UTXOs to use for payment
	Utxos []*Utxo
	// Listings is the list of token listings to create
	Listings []*struct {
		// PayAddress is the address to receive payment when the listing is purchased
		PayAddress string
		// Price is the price in satoshis for the token
		Price uint64
		// ListingUtxo is the UTXO containing the token
		ListingUtxo *TokenUtxo
		// OrdAddress is the address that owns the token
		OrdAddress string
	}
	// PaymentPk is the private key for the payment UTXOs
	PaymentPk *ec.PrivateKey
	// OrdPk is the private key for the token UTXOs
	OrdPk *ec.PrivateKey
	// ChangeAddress is the address to send change to
	ChangeAddress string
	// SatsPerKb is the fee rate in satoshis per kilobyte
	SatsPerKb uint64
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

// PurchaseOrdTokenListingConfig represents configuration for purchasing a token listing
type PurchaseOrdTokenListingConfig struct {
	// Protocol is the token protocol (e.g., TokenTypeBSV21)
	Protocol TokenType
	// TokenID is the ID of the token
	TokenID string
	// Utxos are the UTXOs to use for payment
	Utxos []*Utxo
	// PaymentPk is the private key for the payment UTXOs
	PaymentPk *ec.PrivateKey
	// ListingUtxo is the UTXO containing the token listing
	ListingUtxo *TokenUtxo
	// OrdAddress is the address to send the token to
	OrdAddress string
	// ChangeAddress is the address to send change to
	ChangeAddress string
	// SatsPerKb is the fee rate in satoshis per kilobyte
	SatsPerKb uint64
	// AdditionalPayments is an optional list of additional payments to make
	AdditionalPayments []*PayToAddress
	// Metadata is optional MAP protocol metadata to include in the transfer output
	Metadata map[string][]byte
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

// CancelOrdTokenListingsConfig represents configuration for cancelling token listings
type CancelOrdTokenListingsConfig struct {
	// Utxos are the UTXOs to use for payment
	Utxos []*Utxo
	// ListingUtxos is the list of token listing UTXOs to cancel
	ListingUtxos []*TokenUtxo
	// OrdPk is the private key that owns the tokens
	OrdPk *ec.PrivateKey
	// PaymentPk is the private key for the payment UTXOs
	PaymentPk *ec.PrivateKey
	// ChangeAddress is the address to send change to
	ChangeAddress string
	// SatsPerKb is the fee rate in satoshis per kilobyte
	SatsPerKb uint64
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

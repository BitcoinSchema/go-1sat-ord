# go-1sat-ord

A Go package for creating and managing 1Sat Ordinal inscriptions and transactions using bsv-blockchain/go-sdk and bitcoin-sv/go-templates.

## Description

This library provides functionality for working with Bitcoin SV ordinals, including:

- Creating ordinal inscriptions
- Sending ordinals to new addresses
- Managing regular UTXOs for payments
- BSV21 token deployment and transfers
- Ordinal marketplace functionality (listing, purchasing, canceling)
- Burning ordinals (removing them from circulation)
- Helper functions for fetching UTXOs from APIs
- Utility functions for token UTXO selection and subtype data validation

## Installation

```bash
go get github.com/bitcoinschema/go-1sat-ord
```

## Usage

```go
import (
    "github.com/bitcoinschema/go-1sat-ord"
    "github.com/bsv-blockchain/go-sdk/primitives/ec"
)

// Create a private key for payment
paymentPk, _ := ec.NewPrivateKey()
// Create a private key for ordinals
ordPk, _ := ec.NewPrivateKey()
```

### Create Ordinals

```go
// Prepare utxos to use for transaction fee
utxos := []*ordinals.Utxo{
    {
        TxID:         "txid",
        Vout:         0,
        ScriptPubKey: "script",
        Satoshis:     100000,
    },
}

// Create a destination with inscription
destination := &ordinals.Destination{
    Address: "destination_address",
    File: &ordinals.File{
        Content:     []byte("Hello, world!"),
        ContentType: "text/plain",
    },
    // Optional MAP protocol metadata
    Metadata: map[string][]byte{
        "key": []byte("value"),
    },
}

// If you want to omit metadata from this inscription
// This creates a smaller output with just the P2PKH script
// destination.SetOmitMetadata(true)

// Configure the inscription
config := &ordinals.CreateOrdinalsConfig{
    Utxos:        utxos,
    Destinations: []*ordinals.Destination{destination},
    PaymentPk:    paymentPk,
    ChangeAddress: "change_address",
    // Optional additional payments
    AdditionalPayments: []*ordinals.PayToAddress{
        {
            Address:  "payment_address",
            Satoshis: 5000,
        },
    },
}

// Create the transaction
tx, err := ordinals.CreateOrdinals(config)
if err != nil {
    // Handle error
}

// Broadcast the transaction
result, err := tx.Broadcast(ordinals.OneSatBroadcaster())
if err != nil {
    // Handle error
}
```

### Send Ordinals

```go
// Create a destination for the ordinal
destination := &ordinals.Destination{
    Address: "destination_address",
    // Optional: include a new file to reinscribe
    File: &ordinals.File{
        Content:     []byte("New inscription content"),
        ContentType: "text/plain",
    },
    // Optional: include metadata
    Metadata: map[string][]byte{
        "app":  []byte("myapp"),
        "type": []byte("ord"),
    },
}

// If you want to omit metadata from this inscription
// destination.SetOmitMetadata(true)

// Configure the send
config := &ordinals.SendOrdinalsConfig{
    PaymentUtxos:  paymentUtxos,
    Ordinals:      ordinalUtxos,
    PaymentPk:     paymentPk,
    OrdPk:         ordPk,
    Destinations:  []*ordinals.Destination{destination},
    ChangeAddress: "change_address",
    // Optional: Control whether number of destinations must match number of ordinals
    // Default is true, set to false to allow different counts
    EnforceUniformSend: true,
    // Optional: Include additional payments in the transaction
    AdditionalPayments: []*ordinals.PayToAddress{
        {
            Address:  "payment_address",
            Satoshis: 5000,
        },
    },
}

// Create the transaction
tx, err := ordinals.SendOrdinals(config)
if err != nil {
    // Handle error
}

// Broadcast the transaction
result, err := tx.Broadcast(ordinals.OneSatBroadcaster())
if err != nil {
    // Handle error
}
```

### Deploy BSV21 Token

```go
// Configure the token deployment
config := &ordinals.DeployBsv21TokenConfig{
    Symbol: "TOKEN",
    Icon:   "icon_id",
    Utxos:  utxos,
    InitialDistribution: &ordinals.TokenDistribution{
        Address: "destination_address",
        Tokens:  1000,
    },
    PaymentPk:          paymentPk,
    DestinationAddress: "destination_address",
    ChangeAddress:      "change_address",
}

// Create the transaction
tx, err := ordinals.DeployBsv21Token(config)
if err != nil {
    // Handle error
}

// Broadcast the transaction
result, err := tx.Broadcast(ordinals.OneSatBroadcaster())
if err != nil {
    // Handle error
}
```

### Transfer BSV21 Tokens with Split Configuration

```go
// Set minimum tokens per split output
threshold := 100.0

// Configure token transfer with split configuration
config := &ordinals.TransferBsv21TokenConfig{
    Protocol:    ordinals.TokenTypeBSV21,
    TokenID:     "your-token-id",
    Utxos:       paymentUtxos,
    InputTokens: tokenUtxos,
    Distributions: []*ordinals.TokenDistribution{
        {
            Address: "recipient-address",
            Tokens:  100,
            // Optional: Omit metadata from this distribution output
            // This creates a smaller output with just the P2PKH script
            OmitMetadata: true,
        },
    },
    PaymentPk:     paymentPk,
    OrdPk:         ordPk,
    ChangeAddress: "change-address",
    TokenInputMode: ordinals.TokenInputModeNeeded, // Or TokenInputModeAll
    // Split configuration for token change outputs
    SplitConfig: &ordinals.TokenSplitConfig{
        // Number of outputs to split the token change into
        Outputs: 3,
        // Minimum amount of tokens per output (optional)
        Threshold: &threshold, // where threshold is float64
        // Omit metadata from change outputs (optional)
        // This creates smaller outputs with just the P2PKH script
        OmitMetadata: true,
    },
}

// Create the transaction
tx, err := ordinals.TransferOrdTokens(config)
if err != nil {
    // Handle error
}

// Broadcast the transaction
result, err := tx.Broadcast(ordinals.OneSatBroadcaster())
if err != nil {
    // Handle error
}
```

### Burn Ordinals

```go
// Configure the ordinal burning
config := &ordinals.BurnOrdinalsConfig{
    // Payment UTXOs for transaction fees
    PaymentUtxos: paymentUtxos,
    PaymentPk:    paymentPk,
    // Ordinals to burn
    Ordinals: nftUtxos,
    OrdPk:    ordPk,
    // Change address for leftover satoshis from payment utxos
    ChangeAddress: "change_address",
    // Optional MAP protocol metadata
    Metadata: map[string][]byte{
        "app":  []byte("myapp"),
        "type": []byte("ord"),
        "op":   []byte("burn"),
    },
}

// Create the transaction
tx, err := ordinals.BurnOrdinals(config)
if err != nil {
    // Handle error
}

// Broadcast the transaction
result, err := tx.Broadcast(ordinals.OneSatBroadcaster())
if err != nil {
    // Handle error
}
```

### Token Marketplace Functions

#### Create Token Listings

```go
// Configure the token listings
config := &ordinals.CreateOrdTokenListingsConfig{
    Utxos: paymentUtxos,
    Listings: []*struct {
        PayAddress  string
        Price       uint64
        ListingUtxo *ordinals.TokenUtxo
        OrdAddress  string
    }{
        {
            PayAddress:  "seller_payment_address",
            Price:       1000000, // Price in satoshis
            ListingUtxo: tokenUtxo,
            OrdAddress:  "seller_address",
        },
    },
    PaymentPk:     paymentPk,
    OrdPk:         ordPk,
    ChangeAddress: "change_address",
}

// Create the transaction
tx, err := ordinals.CreateOrdTokenListings(config)
if err != nil {
    // Handle error
}

// Broadcast the transaction
result, err := tx.Broadcast(ordinals.OneSatBroadcaster())
if err != nil {
    // Handle error
}
```

#### Purchase Token Listing

```go
// Configure the token purchase
config := &ordinals.PurchaseOrdTokenListingConfig{
    Protocol:    ordinals.TokenTypeBSV21,
    TokenID:     "token_id",
    Utxos:       paymentUtxos,
    PaymentPk:   paymentPk,
    ListingUtxo: tokenUtxo,
    OrdAddress:  "buyer_address",
    // Optional additional payments
    AdditionalPayments: []*ordinals.PayToAddress{
        {
            Address:  "fee_recipient_address",
            Satoshis: 10000,
        },
    },
    ChangeAddress: "change_address",
}

// Create the transaction
tx, err := ordinals.PurchaseOrdTokenListing(config)
if err != nil {
    // Handle error
}

// Broadcast the transaction
result, err := tx.Broadcast(ordinals.OneSatBroadcaster())
if err != nil {
    // Handle error
}
```

#### Cancel Token Listings

```go
// Configure the token listing cancellation
config := &ordinals.CancelOrdTokenListingsConfig{
    Utxos:        paymentUtxos,
    ListingUtxos: tokenListingUtxos,
    OrdPk:        ordPk,
    PaymentPk:    paymentPk,
    ChangeAddress: "change_address",
}

// Create the transaction
tx, err := ordinals.CancelOrdTokenListings(config)
if err != nil {
    // Handle error
}

// Broadcast the transaction
result, err := tx.Broadcast(ordinals.OneSatBroadcaster())
if err != nil {
    // Handle error
}
```

### Helper Functions

#### Fetch UTXOs

```go
// Fetch UTXOs for payment
paymentUtxos, err := ordinals.FetchPayUtxos("your-payment-address")
if err != nil {
    // Handle error
}

// Fetch NFT UTXOs
nftUtxos, err := ordinals.FetchNftUtxos("your-nft-address", "collection-id")
if err != nil {
    // Handle error
}

// Fetch token UTXOs
tokenUtxos, err := ordinals.FetchTokenUtxos(ordinals.TokenTypeBSV21, "token-id", "your-token-address")
if err != nil {
    // Handle error
}
```

#### Select Token UTXOs

```go
// Define selection options
options := &ordinals.TokenSelectionOptions{
    InputStrategy:  ordinals.TokenSelectionStrategyLargestFirst,
    OutputStrategy: ordinals.TokenSelectionStrategySmallestFirst,
}

// Select token UTXOs
result := ordinals.SelectTokenUtxos(tokenUtxos, 10.5, 8, options)
if !result.IsEnough {
    // Not enough tokens available
    // Handle insufficient funds
} else {
    // Use the selected UTXOs
    selectedUtxos := result.SelectedUtxos
    totalAmount := result.TotalSelected
    // Continue with transaction
}
```

#### Validate Subtype Data

```go
// Create subtype data for a collection
collectionData := &ordinals.SubTypeData{
    Description: "My Collection",
    Quantity:    100,
    Traits: map[string]interface{}{
        "category": "art",
        "rarity":   "common",
    },
}

// Validate collection data
result := ordinals.ValidateSubTypeData(ordinals.TokenTypeBSV21, "collection", collectionData)
if !result.Valid {
    // Handle validation errors
    for _, err := range result.Errors {
        fmt.Println("Validation error:", err)
    }
} else {
    // Data is valid, proceed with collection creation
}

// Create subtype data for a collection item
itemData := &ordinals.SubTypeData{
    CollectionID: "collection123",
    MintNumber:   1,
    Rank:         5,
    RarityLabel:  "rare",
}

// Validate collection item data
result = ordinals.ValidateSubTypeData(ordinals.TokenTypeBSV21, "collectionItem", itemData)
if !result.Valid {
    // Handle validation errors
} else {
    // Data is valid, proceed with item creation
}
```

## Testing

The library includes comprehensive tests for all functionality, including token selection strategies, marketplace operations, and token transfers with various configurations. To run the tests:

```bash
go test -v ./...
```

## Error Handling

The library provides detailed error messages for common issues:

- Insufficient funds
- Invalid addresses
- Script creation errors
- Token validation failures
- UTXO fetch failures

All errors are wrapped with context to help identify the source of the problem.

## More Information

[1Sat Ordinals](https://github.com/bitcoinschema/1sat-ordinals)

## Token Functions

- `DeployToken`: Create a new BSV-20 or BSV-21 token deployment transaction
- `MintBsv20`: Mint new BSV-20 tokens
- `TransferOrdTokens`: Transfer BSV-20 or BSV-21 tokens with advanced distribution options
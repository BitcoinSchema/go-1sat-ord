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

// Configure the inscription
config := &ordinals.CreateOrdinalsConfig{
    Utxos: utxos,
    Destinations: []*struct {
        Address     string
        File        *ordinals.File
        Metadata    map[string][]byte
    }{
        {
            Address: "destination_address",
            File: &ordinals.File{
                Content:     []byte("Hello, world!"),
                ContentType: "text/plain",
            },
            // Optional MAP protocol metadata
            Metadata: map[string][]byte{
                "key": []byte("value"),
            },
        },
    },
    PaymentPk:     paymentPk,
    ChangeAddress: "change_address",
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
// Configure the send
config := &ordinals.SendOrdinalsConfig{
    PaymentUtxos: paymentUtxos,
    Ordinals:     ordinalUtxos,
    PaymentPk:    paymentPk,
    OrdPk:        ordPk,
    Destinations: []*struct {
        Address  string
        File     *ordinals.File
        Metadata map[string][]byte
    }{
        {
            Address: "destination_address",
        },
    },
    ChangeAddress: "change_address",
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
    TokenID:     "token_id",
    Utxos:       paymentUtxos,
    InputTokens: tokenUtxos,
    Distributions: []*ordinals.TokenDistribution{
        {
            Address: "destination_address",
            Tokens:  500,
        },
    },
    PaymentPk:     paymentPk,
    OrdPk:         ordPk,
    ChangeAddress: "change_address",
    // Use TokenInputModeNeeded to consume only the needed tokens
    TokenInputMode: ordinals.TokenInputModeNeeded,
    // Configure token splitting for change outputs
    SplitConfig: &ordinals.TokenSplitConfig{
        // Split into 3 outputs
        Outputs: 3,
        // Minimum tokens per output
        Threshold: &threshold,
        // Omit metadata from change outputs to reduce size
        OmitMetadata: true,
    },
}

// Create the transaction
tx, err := ordinals.TransferOrdToken(config)
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

### Helper Functions

```go
// Fetch UTXOs for payment
utxos, err := ordinals.FetchPayUtxos("address")
if err != nil {
    // Handle error
}

// Fetch NFT UTXOs
nftUtxos, err := ordinals.FetchNftUtxos("address", "collection_id")
if err != nil {
    // Handle error
}

// Fetch Token UTXOs
tokenUtxos, err := ordinals.FetchTokenUtxos(ordinals.TokenTypeBSV21, "token_id", "address")
if err != nil {
    // Handle error
}
```

## More Information

[1Sat Ordinals](https://github.com/bitcoinschema/1sat-ordinals)
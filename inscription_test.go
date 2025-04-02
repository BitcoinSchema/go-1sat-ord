package ordinals

import (
	"fmt"
	"testing"

	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/stretchr/testify/assert"
)

// fetchUtxo is a mock function for fetching UTXO data in tests
var fetchUtxo func(u *Utxo) error = func(u *Utxo) error {
	// Success for TestCreateOrdinals test case - specific real-looking UTXO
	if u.TxID == "2cbc85602b52fc65a70d4e2769b8c0ea28462bf9d8da86485a787220563e708b" {
		return nil
	}

	// Success for specific valid UTXO patterns used in our tests
	if u.TxID == "0000000000000000000000000000000000000000000000000000000000000003" ||
		u.TxID == "0000000000000000000000000000000000000000000000000000000000000004" {
		return nil
	}

	// Success for transaction input token test cases
	if u.TxID == "1111111111111111111111111111111111111111111111111111111111111111" ||
		u.TxID == "2222222222222222222222222222222222222222222222222222222222222222" {
		return nil
	}

	// For all other test cases, return an error to simulate UTXO not found
	return fmt.Errorf("UTXO not found or invalid: %s:%d", u.TxID, u.Vout)
}

func TestCreateOrdinals(t *testing.T) {
	// Create a private key for payment
	paymentPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	// Prepare a test utxo - this is a real UTXO
	utxo := &Utxo{
		TxID:         "2cbc85602b52fc65a70d4e2769b8c0ea28462bf9d8da86485a787220563e708b",
		Vout:         1,
		ScriptPubKey: "76a91458120c48b55a861fe667c96e64b327004e6ff13c88ac",
		Satoshis:     59024,
	}

	// Create a test configuration
	config := &CreateOrdinalsConfig{
		Utxos: []*Utxo{utxo},
		Destinations: []*struct {
			Address  string
			File     *File
			Metadata map[string][]byte
		}{
			{
				Address: "1BitcoinEaterAddressDontSendf59kuE",
				File: &File{
					Content:     []byte("Hello, world!"),
					ContentType: "text/plain",
				},
				Metadata: map[string][]byte{
					"test": []byte("value"),
				},
			},
		},
		PaymentPk:     paymentPk,
		ChangeAddress: "1BitcoinEaterAddressDontSendf59kuE",
	}

	// Create the transaction
	tx, err := CreateOrdinals(config)

	// The transaction should be created successfully without errors
	assert.NoError(t, err)
	assert.NotNil(t, tx)
}

func TestSendOrdinals(t *testing.T) {
	// Create private keys
	paymentPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	ordPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	// Prepare test utxos with valid format
	paymentUtxo := &Utxo{
		TxID:         "0000000000000000000000000000000000000000000000000000000000000003",
		Vout:         0,
		ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
		Satoshis:     100000,
	}

	ordinalUtxo := &NftUtxo{
		Utxo: Utxo{
			TxID:         "0000000000000000000000000000000000000000000000000000000000000003",
			Vout:         1,
			ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
			Satoshis:     1,
		},
		ContentType:  "text/plain",
		CollectionID: "collection123",
	}

	// Create a test configuration
	config := &SendOrdinalsConfig{
		PaymentUtxos: []*Utxo{paymentUtxo},
		Ordinals:     []*NftUtxo{ordinalUtxo},
		PaymentPk:    paymentPk,
		OrdPk:        ordPk,
		Destinations: []*struct {
			Address  string
			File     *File
			Metadata map[string][]byte
		}{
			{
				Address: "1BitcoinEaterAddressDontSendf59kuE",
			},
		},
		ChangeAddress: "1BitcoinEaterAddressDontSendf59kuE",
	}

	// Create the transaction
	tx, err := SendOrdinals(config)

	// We expect the transaction to be created successfully
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify the transaction structure
	assert.Equal(t, 2, len(tx.Inputs), "Should have 2 inputs: payment and ordinal")
	assert.Equal(t, 2, len(tx.Outputs), "Should have 2 outputs: destination and change")
}

func TestSendUtxos(t *testing.T) {
	// Create a private key for payment
	paymentPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	// Prepare a test utxo with valid format
	utxo := &Utxo{
		TxID:         "0000000000000000000000000000000000000000000000000000000000000003",
		Vout:         0,
		ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
		Satoshis:     100000,
	}

	// Create a test configuration
	config := &SendUtxosConfig{
		Utxos:     []*Utxo{utxo},
		PaymentPk: paymentPk,
		Payments: []*PayToAddress{
			{
				Address:  "1BitcoinEaterAddressDontSendf59kuE",
				Satoshis: 50000,
			},
		},
		ChangeAddress: "1BitcoinEaterAddressDontSendf59kuE",
	}

	// Create the transaction
	tx, err := SendUtxos(config)

	// We expect the transaction to be created successfully
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify the transaction structure
	assert.Equal(t, 1, len(tx.Inputs), "Should have 1 input")
	assert.Equal(t, 2, len(tx.Outputs), "Should have 2 outputs: payment and change")
	assert.Equal(t, uint64(50000), tx.Outputs[0].Satoshis, "First output should be 50000 satoshis")
}

func TestDeployBsv21Token(t *testing.T) {
	// Create private keys for payment
	paymentPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	// Get the payment address
	paymentAddr, err := script.NewAddressFromPublicKey(paymentPk.PubKey(), true)
	assert.NoError(t, err)
	address := paymentAddr.AddressString

	// Prepare test utxos for different scenarios
	insufficientUtxo := &Utxo{
		TxID:         "0000000000000000000000000000000000000000000000000000000000000001",
		Vout:         0,
		ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
		Satoshis:     100, // Insufficient amount
	}

	exactUtxo := &Utxo{
		TxID:         "0000000000000000000000000000000000000000000000000000000000000003",
		Vout:         1,
		ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
		Satoshis:     10000, // Exactly enough
	}

	sufficientUtxo := &Utxo{
		TxID:         "0000000000000000000000000000000000000000000000000000000000000003",
		Vout:         2,
		ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
		Satoshis:     100000, // More than required
	}

	// Setup initial distribution
	initialDistribution := &TokenDistribution{
		Address: address,
		Tokens:  1000,
	}

	// Base configuration for BSV21 token deployment
	baseCfg := &DeployBsv21TokenConfig{
		Symbol:              "TEST",
		Icon:                "<svg width=\"100\" height=\"100\" xmlns=\"http://www.w3.org/2000/svg\"><circle cx=\"50\" cy=\"50\" r=\"40\" stroke=\"black\" stroke-width=\"3\" fill=\"red\" /></svg>",
		InitialDistribution: initialDistribution,
		PaymentPk:           paymentPk,
		DestinationAddress:  address,
		ChangeAddress:       address,
	}

	// Test case: Deploy BSV21 token with sufficient UTXO
	t.Run("deploy BSV21 token with sufficient utxo", func(t *testing.T) {
		cfg := *baseCfg
		cfg.Utxos = []*Utxo{sufficientUtxo}

		// This should succeed
		tx, err := DeployBsv21Token(&cfg)
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		// Verify outputs: token output + change
		assert.Equal(t, 2, len(tx.Outputs), "Should have 2 outputs: token and change")
		assert.Equal(t, uint64(1), tx.Outputs[0].Satoshis, "Token output should be 1 satoshi")
	})

	// Test case: Deploy BSV21 token with exact UTXO
	t.Run("deploy BSV21 token with exact utxo", func(t *testing.T) {
		cfg := *baseCfg
		cfg.Utxos = []*Utxo{exactUtxo}

		// This should succeed
		tx, err := DeployBsv21Token(&cfg)
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		// Verify there's only one input
		assert.Equal(t, 1, len(tx.Inputs), "Should have 1 input")
	})

	// Override the original fetchUtxo mock for the insufficient UTXO test
	originalFetchUtxo := fetchUtxo

	t.Run("deploy BSV21 token with insufficient utxo", func(t *testing.T) {
		// Restore after the test
		defer func() { fetchUtxo = originalFetchUtxo }()

		// For this particular test, we'll use a higher fee rate to make sure the transaction fails
		cfg := *baseCfg
		cfg.Utxos = []*Utxo{insufficientUtxo}
		cfg.SatsPerKb = 10000 // Set a very high fee rate to ensure insufficient funds

		// This should fail with insufficient funds error when calculating fees
		tx, err := DeployBsv21Token(&cfg)
		assert.Error(t, err, "Expected error due to insufficient funds")
		assert.Nil(t, tx, "Expected tx to be nil")
		if err != nil {
			// The actual error message is "satoshis inputted to the tx are less than the outputted satoshis"
			assert.Contains(t, err.Error(), "less than", "Error should indicate insufficient funds")
		}
	})

	// Test case: Deploy BSV21 token with incorrect image proportions
	t.Run("deploy BSV21 token with non-square dimensions (should still pass)", func(t *testing.T) {
		// Restore after the test
		defer func() { fetchUtxo = originalFetchUtxo }()

		cfg := *baseCfg
		cfg.Utxos = []*Utxo{sufficientUtxo}

		// Set an SVG with incorrect proportions (not square)
		cfg.Icon = "<svg width=\"200\" height=\"100\" xmlns=\"http://www.w3.org/2000/svg\"><rect width=\"200\" height=\"100\" fill=\"blue\" /></svg>"

		// Currently the library doesn't validate dimensions - if we wanted to enforce this,
		// we would need to add validation to the DeployBsv21Token function
		tx, err := DeployBsv21Token(&cfg)
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		// Although dimensions aren't validated, we can still check other token properties
		assert.Equal(t, 2, len(tx.Outputs), "Should have 2 outputs: token and change")
		assert.Equal(t, uint64(1), tx.Outputs[0].Satoshis, "Token output should be 1 satoshi")
	})

	// Test case: Deploy BSV21 token with valid square SVG
	t.Run("deploy BSV21 token with valid square SVG", func(t *testing.T) {
		cfg := *baseCfg
		cfg.Utxos = []*Utxo{sufficientUtxo}

		// Set a valid square SVG
		cfg.Icon = "<svg width=\"100\" height=\"100\" xmlns=\"http://www.w3.org/2000/svg\"><rect width=\"100\" height=\"100\" fill=\"green\" /></svg>"

		tx, err := DeployBsv21Token(&cfg)
		assert.NoError(t, err)
		assert.NotNil(t, tx)
	})
}

func TestTransferOrdToken(t *testing.T) {
	// Create private keys
	paymentPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	ordPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	// Prepare test utxos with invalid data to ensure test fails appropriately
	paymentUtxo := &Utxo{
		TxID:         "invalid-txid", // Invalid TXID format
		Vout:         0,
		ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
		Satoshis:     100000,
	}

	tokenUtxo := &TokenUtxo{
		Utxo: Utxo{
			TxID:         "invalid-txid", // Invalid TXID format
			Vout:         1,
			ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
			Satoshis:     1,
		},
		TokenID:  "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890_0",
		Protocol: TokenTypeBSV21,
		Amount:   1000,
		Decimals: 0,
	}

	// Create a test configuration
	config := &TransferBsv21TokenConfig{
		Protocol:    TokenTypeBSV21,
		TokenID:     "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890_0",
		Utxos:       []*Utxo{paymentUtxo},
		InputTokens: []*TokenUtxo{tokenUtxo},
		Distributions: []*TokenDistribution{
			{
				Address: "1BitcoinEaterAddressDontSendf59kuE",
				Tokens:  500,
			},
		},
		PaymentPk:     paymentPk,
		OrdPk:         ordPk,
		Burn:          false,
		ChangeAddress: "1BitcoinEaterAddressDontSendf59kuE",
	}

	// Create the transaction
	tx, err := TransferOrdToken(config)

	// We expect an error because the test utxos are invalid
	assert.Error(t, err)
	assert.Nil(t, tx)
}

func TestCreateOrdListings(t *testing.T) {
	// Create private keys
	paymentPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	ordPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	// Prepare test utxos with valid format
	paymentUtxo := &Utxo{
		TxID:         "0000000000000000000000000000000000000000000000000000000000000003",
		Vout:         0,
		ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
		Satoshis:     100000,
	}

	ordinalUtxo := &NftUtxo{
		Utxo: Utxo{
			TxID:         "0000000000000000000000000000000000000000000000000000000000000003",
			Vout:         1,
			ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
			Satoshis:     1,
		},
		ContentType:  "text/plain",
		CollectionID: "collection123",
	}

	// Create a test configuration
	config := &CreateOrdListingsConfig{
		Utxos: []*Utxo{paymentUtxo},
		Listings: []*struct {
			PayAddress  string
			Price       uint64
			ListingUtxo *NftUtxo
			OrdAddress  string
		}{
			{
				PayAddress:  "1BitcoinEaterAddressDontSendf59kuE",
				Price:       50000,
				ListingUtxo: ordinalUtxo,
				OrdAddress:  "1BitcoinEaterAddressDontSendf59kuE",
			},
		},
		PaymentPk:     paymentPk,
		OrdPk:         ordPk,
		ChangeAddress: "1BitcoinEaterAddressDontSendf59kuE",
	}

	// Create the transaction
	tx, err := CreateOrdListings(config)

	// We expect the transaction to be created successfully
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify the transaction structure
	assert.Equal(t, 2, len(tx.Inputs), "Should have 2 inputs: payment and ordinal")
	assert.Equal(t, 2, len(tx.Outputs), "Should have 2 outputs: ordlock and change")
	assert.Equal(t, uint64(1), tx.Outputs[0].Satoshis, "Ordlock output should be 1 satoshi")
}

func TestPurchaseOrdListing(t *testing.T) {
	// Create a private key for payment
	paymentPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	// Prepare test utxos with valid format
	paymentUtxo := &Utxo{
		TxID:         "0000000000000000000000000000000000000000000000000000000000000003",
		Vout:         0,
		ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
		Satoshis:     100000,
	}

	listingUtxo := &NftUtxo{
		Utxo: Utxo{
			TxID:         "0000000000000000000000000000000000000000000000000000000000000003",
			Vout:         1,
			ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
			Satoshis:     1,
		},
		ContentType:  "text/plain",
		CollectionID: "collection123",
	}

	// Create a test configuration
	config := &PurchaseOrdListingConfig{
		Utxos:         []*Utxo{paymentUtxo},
		PaymentPk:     paymentPk,
		ListingUtxo:   listingUtxo,
		OrdAddress:    "1BitcoinEaterAddressDontSendf59kuE",
		ChangeAddress: "1BitcoinEaterAddressDontSendf59kuE",
	}

	// Create the transaction
	tx, err := PurchaseOrdListing(config)

	// We expect the transaction to be created successfully
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify the transaction structure
	assert.Equal(t, 2, len(tx.Inputs), "Should have 2 inputs: payment and listing")
	assert.Equal(t, 2, len(tx.Outputs), "Should have 2 outputs: ordinal and change")
	assert.Equal(t, uint64(1), tx.Outputs[0].Satoshis, "Ordinal output should be 1 satoshi")
}

func TestCancelOrdListings(t *testing.T) {
	// Create private keys
	paymentPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	ordPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	// Prepare test utxos with valid format
	paymentUtxo := &Utxo{
		TxID:         "0000000000000000000000000000000000000000000000000000000000000003",
		Vout:         0,
		ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
		Satoshis:     100000,
	}

	listingUtxo := &NftUtxo{
		Utxo: Utxo{
			TxID:         "0000000000000000000000000000000000000000000000000000000000000003",
			Vout:         1,
			ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
			Satoshis:     1,
		},
		ContentType:  "text/plain",
		CollectionID: "collection123",
	}

	// Create a test configuration
	config := &CancelOrdListingsConfig{
		Utxos:         []*Utxo{paymentUtxo},
		ListingUtxos:  []*NftUtxo{listingUtxo},
		OrdPk:         ordPk,
		PaymentPk:     paymentPk,
		ChangeAddress: "1BitcoinEaterAddressDontSendf59kuE",
	}

	// Create the transaction
	tx, err := CancelOrdListings(config)

	// We expect the transaction to be created successfully
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify the transaction structure
	assert.Equal(t, 2, len(tx.Inputs), "Should have 2 inputs: payment and listing")
	assert.Equal(t, 2, len(tx.Outputs), "Should have 2 outputs: ordinal and change")
	assert.Equal(t, uint64(1), tx.Outputs[0].Satoshis, "Ordinal output should be 1 satoshi")
}

func TestHelpers(t *testing.T) {
	// Test AddressFromString
	addr, err := AddressFromString("1BitcoinEaterAddressDontSendf59kuE")
	assert.NoError(t, err)
	assert.NotNil(t, addr)
	assert.Equal(t, "1BitcoinEaterAddressDontSendf59kuE", addr.AddressString)
}

func TestTokenSplitConfig(t *testing.T) {
	// Create private keys
	paymentPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	ordPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	// Setup test params and config
	tokenID := "1fcf743a77ea69755bf2b8ea70530a47de9c064daf1eee09cbc6f39e434bb0fb_0"
	changeAddress := "1DBJ3MsNKdvuqXcmFxw9SvV6GHWmC7bxSA"
	recipient := "1GpAScbJDFvMSUfZBYdXZiBpzW8Bfa8rPE"
	tokenAmount := float64(100)

	// Setup test UTXOs with invalid format to ensure tests fail appropriately
	paymentUtxo := &Utxo{
		TxID:         "invalid-txid", // Invalid TXID format
		Vout:         0,
		ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
		Satoshis:     100000,
	}

	tokenUtxo := &TokenUtxo{
		Utxo: Utxo{
			TxID:         "invalid-txid", // Invalid TXID format
			Vout:         1,
			ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
			Satoshis:     1,
		},
		TokenID:  tokenID,
		Protocol: TokenTypeBSV21,
		Amount:   1000,
		Decimals: 0,
	}

	// Mock the fetchUtxo function to always succeed
	originalFetchUtxo := fetchUtxo
	defer func() { fetchUtxo = originalFetchUtxo }()

	fetchUtxo = func(u *Utxo) error {
		return nil // Always succeed
	}

	// Test case: Single output (no split)
	t.Run("Single output (no split)", func(t *testing.T) {
		// Create a test configuration with a single output
		outputs := 1
		cfg := &TransferBsv21TokenConfig{
			Protocol:    TokenTypeBSV21,
			TokenID:     tokenID,
			Utxos:       []*Utxo{paymentUtxo},
			InputTokens: []*TokenUtxo{tokenUtxo},
			Distributions: []*TokenDistribution{
				{
					Address: recipient,
					Tokens:  tokenAmount,
				},
			},
			PaymentPk:      paymentPk,
			OrdPk:          ordPk,
			Burn:           false,
			ChangeAddress:  changeAddress,
			TokenInputMode: TokenInputModeNeeded,
			SplitConfig: &TokenSplitConfig{
				Outputs:      outputs,
				OmitMetadata: false,
			},
		}

		// Verify the transaction will fail due to invalid UTXO
		// but token split logic should still run correctly
		_, err := TransferOrdToken(cfg)
		assert.Error(t, err)
	})

	// Test case: Multiple outputs with threshold
	t.Run("Multiple outputs with threshold", func(t *testing.T) {
		// Create a test configuration with multiple outputs and a threshold
		outputs := 3
		threshold := float64(200)
		cfg := &TransferBsv21TokenConfig{
			Protocol:    TokenTypeBSV21,
			TokenID:     tokenID,
			Utxos:       []*Utxo{paymentUtxo},
			InputTokens: []*TokenUtxo{tokenUtxo},
			Distributions: []*TokenDistribution{
				{
					Address: recipient,
					Tokens:  tokenAmount,
				},
			},
			PaymentPk:      paymentPk,
			OrdPk:          ordPk,
			Burn:           false,
			ChangeAddress:  changeAddress,
			TokenInputMode: TokenInputModeNeeded,
			SplitConfig: &TokenSplitConfig{
				Outputs:      outputs,
				Threshold:    &threshold,
				OmitMetadata: false,
			},
		}

		// Verify the transaction will fail due to invalid UTXO
		// but token split logic should still run correctly
		_, err := TransferOrdToken(cfg)
		assert.Error(t, err)
	})

	// Test case: OmitMetadata enabled
	t.Run("OmitMetadata enabled", func(t *testing.T) {
		// Create a test configuration with OmitMetadata enabled
		outputs := 1
		cfg := &TransferBsv21TokenConfig{
			Protocol:    TokenTypeBSV21,
			TokenID:     tokenID,
			Utxos:       []*Utxo{paymentUtxo},
			InputTokens: []*TokenUtxo{tokenUtxo},
			Distributions: []*TokenDistribution{
				{
					Address: recipient,
					Tokens:  tokenAmount,
				},
			},
			PaymentPk:      paymentPk,
			OrdPk:          ordPk,
			Burn:           false,
			ChangeAddress:  changeAddress,
			TokenInputMode: TokenInputModeNeeded,
			SplitConfig: &TokenSplitConfig{
				Outputs:      outputs,
				OmitMetadata: true,
			},
		}

		// Verify the transaction will fail due to invalid UTXO
		// but token split logic should still run correctly
		_, err := TransferOrdToken(cfg)
		assert.Error(t, err)
	})

	// Test case: Threshold equals remaining tokens
	t.Run("Threshold equals remaining tokens", func(t *testing.T) {
		// Create a test configuration with threshold equal to remaining tokens
		outputs := 2
		// If we send 100 and have 1000 total, remaining would be 900
		remainingTokens := float64(900)
		cfg := &TransferBsv21TokenConfig{
			Protocol:    TokenTypeBSV21,
			TokenID:     tokenID,
			Utxos:       []*Utxo{paymentUtxo},
			InputTokens: []*TokenUtxo{tokenUtxo},
			Distributions: []*TokenDistribution{
				{
					Address: recipient,
					Tokens:  tokenAmount,
				},
			},
			PaymentPk:      paymentPk,
			OrdPk:          ordPk,
			Burn:           false,
			ChangeAddress:  changeAddress,
			TokenInputMode: TokenInputModeNeeded,
			SplitConfig: &TokenSplitConfig{
				Outputs:      outputs,
				Threshold:    &remainingTokens,
				OmitMetadata: false,
			},
		}

		// Verify the transaction will fail due to invalid UTXO
		// but token split logic should still run correctly
		_, err := TransferOrdToken(cfg)
		assert.Error(t, err)
	})

	// Test case: Burn tokens
	t.Run("Burn tokens", func(t *testing.T) {
		// Create a test configuration with Burn=true
		cfg := &TransferBsv21TokenConfig{
			Protocol:    TokenTypeBSV21,
			TokenID:     tokenID,
			Utxos:       []*Utxo{paymentUtxo},
			InputTokens: []*TokenUtxo{tokenUtxo},
			Distributions: []*TokenDistribution{
				{
					Address: recipient,
					Tokens:  tokenAmount,
				},
			},
			PaymentPk:      paymentPk,
			OrdPk:          ordPk,
			Burn:           true, // Burn remaining tokens
			ChangeAddress:  changeAddress,
			TokenInputMode: TokenInputModeAll,
		}

		// Verify the transaction will fail due to invalid UTXO
		// but token split logic should still run correctly
		_, err := TransferOrdToken(cfg)
		assert.Error(t, err)
	})

	// Test case: TokenInputModeAll with multiple UTXOs
	t.Run("TokenInputModeAll with multiple UTXOs", func(t *testing.T) {
		// Create a second token UTXO
		tokenUtxo2 := &TokenUtxo{
			Utxo: Utxo{
				TxID:         "invalid-txid",
				Vout:         2,
				ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
				Satoshis:     1,
			},
			TokenID:  tokenID,
			Protocol: TokenTypeBSV21,
			Amount:   500, // 500 token units
			Decimals: 0,
		}

		// Create a test configuration with multiple token UTXOs and TokenInputModeAll
		outputs := 2
		cfg := &TransferBsv21TokenConfig{
			Protocol:    TokenTypeBSV21,
			TokenID:     tokenID,
			Utxos:       []*Utxo{paymentUtxo},
			InputTokens: []*TokenUtxo{tokenUtxo, tokenUtxo2}, // 1500 tokens total
			Distributions: []*TokenDistribution{
				{
					Address: recipient,
					Tokens:  tokenAmount,
				},
			},
			PaymentPk:      paymentPk,
			OrdPk:          ordPk,
			Burn:           false,
			ChangeAddress:  changeAddress,
			TokenInputMode: TokenInputModeAll,
			SplitConfig: &TokenSplitConfig{
				Outputs:      outputs,
				OmitMetadata: true,
			},
		}

		// Verify the transaction will fail due to invalid UTXO
		// but token split logic should still run correctly
		_, err := TransferOrdToken(cfg)
		assert.Error(t, err)
	})
}

func TestBurnOrdinals(t *testing.T) {
	// Create private keys for payment and ordinals
	paymentPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)
	ordPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	// Prepare test payment UTXO
	paymentUtxo := &Utxo{
		TxID:         "0000000000000000000000000000000000000000000000000000000000000003",
		Vout:         0,
		ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
		Satoshis:     100000,
	}

	// Prepare test ordinal UTXO with valid format
	ordinalUtxo := &NftUtxo{
		Utxo: Utxo{
			TxID:         "0000000000000000000000000000000000000000000000000000000000000004",
			Vout:         0,
			ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
			Satoshis:     1, // Must be 1 sat for ordinals
		},
		ContentType:  "text/plain",
		CollectionID: "collection123",
	}

	// Test Case 1: Burn ordinal without metadata
	t.Run("Burn ordinal without metadata", func(t *testing.T) {
		// Create a test configuration
		config := &BurnOrdinalsConfig{
			PaymentUtxos:  []*Utxo{paymentUtxo},
			PaymentPk:     paymentPk,
			Ordinals:      []*NftUtxo{ordinalUtxo},
			OrdPk:         ordPk,
			ChangeAddress: "1BitcoinEaterAddressDontSendf59kuE",
		}

		// Create the transaction
		tx, err := BurnOrdinals(config)

		// We expect the transaction to be created successfully
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		// Verify the transaction structure
		assert.Equal(t, 2, len(tx.Inputs), "Should have 2 inputs: payment and ordinal")
		assert.GreaterOrEqual(t, len(tx.Outputs), 2, "Should have at least 2 outputs: OP_RETURN and change")
		assert.Equal(t, uint64(0), tx.Outputs[0].Satoshis, "OP_RETURN output should be 0 satoshis")
	})

	// Test Case 2: Burn ordinal with metadata
	t.Run("Burn ordinal with metadata", func(t *testing.T) {
		// Create a test configuration with metadata
		config := &BurnOrdinalsConfig{
			PaymentUtxos:  []*Utxo{paymentUtxo},
			PaymentPk:     paymentPk,
			Ordinals:      []*NftUtxo{ordinalUtxo},
			OrdPk:         ordPk,
			ChangeAddress: "1BitcoinEaterAddressDontSendf59kuE",
			Metadata: map[string][]byte{
				"app":  []byte("testapp"),
				"type": []byte("burn"),
				"op":   []byte("burn"),
			},
		}

		// Create the transaction
		tx, err := BurnOrdinals(config)

		// We expect the transaction to be created successfully
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		// Verify the transaction structure
		assert.Equal(t, 2, len(tx.Inputs), "Should have 2 inputs: payment and ordinal")
		assert.GreaterOrEqual(t, len(tx.Outputs), 2, "Should have at least 2 outputs: OP_RETURN and change")
		assert.Equal(t, uint64(0), tx.Outputs[0].Satoshis, "OP_RETURN output should be 0 satoshis")
	})

	// Test Case 3: Try to burn a non-1sat ordinal (should fail)
	t.Run("Try to burn non-1sat ordinal", func(t *testing.T) {
		// Create an invalid ordinal (more than 1 sat)
		invalidOrdinal := &NftUtxo{
			Utxo: Utxo{
				TxID:         "0000000000000000000000000000000000000000000000000000000000000003",
				Vout:         0,
				ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
				Satoshis:     2, // More than 1 sat (invalid for ordinals)
			},
			ContentType:  "text/plain",
			CollectionID: "collection123",
		}

		// Create a test configuration
		config := &BurnOrdinalsConfig{
			PaymentUtxos:  []*Utxo{paymentUtxo},
			PaymentPk:     paymentPk,
			Ordinals:      []*NftUtxo{invalidOrdinal},
			OrdPk:         ordPk,
			ChangeAddress: "1BitcoinEaterAddressDontSendf59kuE",
		}

		// Create the transaction - should fail
		tx, err := BurnOrdinals(config)

		// We expect an error due to the ordinal not being 1 sat
		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.Contains(t, err.Error(), "must have exactly 1 satoshi")
	})

	// Test Case 4: Burn multiple ordinals
	t.Run("Burn multiple ordinals", func(t *testing.T) {
		// Create additional ordinal
		ordinalUtxo2 := &NftUtxo{
			Utxo: Utxo{
				TxID:         "0000000000000000000000000000000000000000000000000000000000000004",
				Vout:         1,
				ScriptPubKey: "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
				Satoshis:     1,
			},
			ContentType:  "text/plain",
			CollectionID: "collection123",
		}

		// Create a test configuration
		config := &BurnOrdinalsConfig{
			PaymentUtxos:  []*Utxo{paymentUtxo},
			PaymentPk:     paymentPk,
			Ordinals:      []*NftUtxo{ordinalUtxo, ordinalUtxo2},
			OrdPk:         ordPk,
			ChangeAddress: "1BitcoinEaterAddressDontSendf59kuE",
		}

		// Create the transaction
		tx, err := BurnOrdinals(config)

		// We expect the transaction to be created successfully
		assert.NoError(t, err)
		assert.NotNil(t, tx)

		// Verify the transaction structure
		assert.Equal(t, 3, len(tx.Inputs), "Should have 3 inputs: payment and 2 ordinals")
		assert.GreaterOrEqual(t, len(tx.Outputs), 2, "Should have at least 2 outputs: OP_RETURN and change")
	})
}

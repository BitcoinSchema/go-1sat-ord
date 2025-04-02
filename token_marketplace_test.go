package ordinals

import (
	"testing"

	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrdTokenListings(t *testing.T) {
	// Create private keys for testing
	paymentPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)
	ordPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	// Get addresses
	paymentAddr, err := script.NewAddressFromPublicKey(paymentPk.PubKey(), true)
	assert.NoError(t, err)
	ordAddr, err := script.NewAddressFromPublicKey(ordPk.PubKey(), true)
	assert.NoError(t, err)

	// Mock a token UTXO
	tokenUtxo := &TokenUtxo{
		Utxo: Utxo{
			TxID:         "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			Vout:         0,
			ScriptPubKey: "76a914b10d25c5ba3dda4e217524c7f7a6d6c53d2ae85588ac",
			Satoshis:     1,
		},
		TokenID:  "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890_0",
		Protocol: TokenTypeBSV21,
		Amount:   1000,
		Decimals: 0,
	}

	// Mock a payment UTXO
	paymentUtxo := &Utxo{
		TxID:         "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567891",
		Vout:         0,
		ScriptPubKey: "76a914a5f427350ffc9a9f0c02e823ff5c3d77c9846fec88ac",
		Satoshis:     100000,
	}

	// Create configuration
	config := &CreateOrdTokenListingsConfig{
		Utxos: []*Utxo{paymentUtxo},
		Listings: []*struct {
			PayAddress  string
			Price       uint64
			ListingUtxo *TokenUtxo
			OrdAddress  string
		}{
			{
				PayAddress:  paymentAddr.AddressString,
				Price:       10000,
				ListingUtxo: tokenUtxo,
				OrdAddress:  ordAddr.AddressString,
			},
		},
		PaymentPk:     paymentPk,
		OrdPk:         ordPk,
		ChangeAddress: paymentAddr.AddressString,
		SatsPerKb:     500,
	}

	// Create the transaction
	tx, err := CreateOrdTokenListings(config)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify the transaction structure
	assert.Equal(t, 2, len(tx.Inputs))                 // 1 payment input + 1 token input
	assert.GreaterOrEqual(t, len(tx.Outputs), 2)       // At least 1 token output + change
	assert.Equal(t, uint64(1), tx.Outputs[0].Satoshis) // 1 sat for ordinals
}

func TestPurchaseOrdTokenListing(t *testing.T) {
	// Create private keys for testing
	paymentPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	// Get address
	paymentAddr, err := script.NewAddressFromPublicKey(paymentPk.PubKey(), true)
	assert.NoError(t, err)

	// Mock a token listing UTXO
	listingUtxo := &TokenUtxo{
		Utxo: Utxo{
			TxID:         "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			Vout:         0,
			ScriptPubKey: "76a914b10d25c5ba3dda4e217524c7f7a6d6c53d2ae85588ac",
			Satoshis:     1,
		},
		TokenID:  "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890_0",
		Protocol: TokenTypeBSV21,
		Amount:   1000,
		Decimals: 0,
	}

	// Mock a payment UTXO
	paymentUtxo := &Utxo{
		TxID:         "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567891",
		Vout:         0,
		ScriptPubKey: "76a914a5f427350ffc9a9f0c02e823ff5c3d77c9846fec88ac",
		Satoshis:     200000,
	}

	// Create configuration
	config := &PurchaseOrdTokenListingConfig{
		Protocol:    TokenTypeBSV21,
		TokenID:     listingUtxo.TokenID,
		Utxos:       []*Utxo{paymentUtxo},
		PaymentPk:   paymentPk,
		ListingUtxo: listingUtxo,
		OrdAddress:  paymentAddr.AddressString,
		AdditionalPayments: []*PayToAddress{
			{
				Address:  paymentAddr.AddressString,
				Satoshis: 5000,
			},
		},
		ChangeAddress: paymentAddr.AddressString,
		SatsPerKb:     500,
	}

	// Create the transaction
	tx, err := PurchaseOrdTokenListing(config)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify the transaction structure
	assert.Equal(t, 2, len(tx.Inputs))                 // 1 listing input + 1 payment input
	assert.GreaterOrEqual(t, len(tx.Outputs), 3)       // token output + payment output + additional payment + change
	assert.Equal(t, uint64(1), tx.Outputs[0].Satoshis) // 1 sat for ordinals
}

func TestCancelOrdTokenListings(t *testing.T) {
	// Create private keys for testing
	paymentPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)
	ordPk, err := ec.NewPrivateKey()
	assert.NoError(t, err)

	// Get address
	paymentAddr, err := script.NewAddressFromPublicKey(paymentPk.PubKey(), true)
	assert.NoError(t, err)

	// Mock a token listing UTXO
	listingUtxo := &TokenUtxo{
		Utxo: Utxo{
			TxID:         "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
			Vout:         0,
			ScriptPubKey: "76a914b10d25c5ba3dda4e217524c7f7a6d6c53d2ae85588ac",
			Satoshis:     1,
		},
		TokenID:  "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890_0",
		Protocol: TokenTypeBSV21,
		Amount:   1000,
		Decimals: 0,
	}

	// Mock a payment UTXO
	paymentUtxo := &Utxo{
		TxID:         "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567891",
		Vout:         0,
		ScriptPubKey: "76a914a5f427350ffc9a9f0c02e823ff5c3d77c9846fec88ac",
		Satoshis:     10000,
	}

	// Create configuration
	config := &CancelOrdTokenListingsConfig{
		Utxos:         []*Utxo{paymentUtxo},
		ListingUtxos:  []*TokenUtxo{listingUtxo},
		OrdPk:         ordPk,
		PaymentPk:     paymentPk,
		ChangeAddress: paymentAddr.AddressString,
		SatsPerKb:     500,
	}

	// Create the transaction
	tx, err := CancelOrdTokenListings(config)
	assert.NoError(t, err)
	assert.NotNil(t, tx)

	// Verify the transaction structure
	assert.Equal(t, 2, len(tx.Inputs))                 // 1 listing input + 1 payment input
	assert.GreaterOrEqual(t, len(tx.Outputs), 2)       // token output + change
	assert.Equal(t, uint64(1), tx.Outputs[0].Satoshis) // 1 sat for ordinals
}

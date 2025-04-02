package ordinals

import (
	"fmt"

	"github.com/bitcoin-sv/go-templates/template/ordlock"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	fee_model "github.com/bsv-blockchain/go-sdk/transaction/fee_model"
	"github.com/bsv-blockchain/go-sdk/transaction/template/p2pkh"
)

// CreateOrdListings creates a listing using an "Ordinal Lock" script
func CreateOrdListings(config *CreateOrdListingsConfig) (*transaction.Transaction, error) {
	// Create a new transaction
	tx := transaction.NewTransaction()

	// Add inputs
	for _, utxo := range config.Utxos {
		unlocker, err := p2pkh.Unlock(config.PaymentPk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create unlocker: %w", err)
		}

		err = tx.AddInputFrom(
			utxo.TxID,
			utxo.Vout,
			utxo.ScriptPubKey,
			utxo.Satoshis,
			unlocker,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add input: %w", err)
		}
	}

	// Add ordinal input (for each listing)
	for _, listing := range config.Listings {
		ordUtxo := listing.ListingUtxo

		unlocker, err := p2pkh.Unlock(config.OrdPk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create ordinal unlocker: %w", err)
		}

		err = tx.AddInputFrom(
			ordUtxo.TxID,
			ordUtxo.Vout,
			ordUtxo.ScriptPubKey,
			ordUtxo.Satoshis,
			unlocker,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add ordinal input: %w", err)
		}

		// Create seller address (for return on cancel)
		sellerAddr, err := script.NewAddressFromString(listing.OrdAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to create seller address: %w", err)
		}

		// Create pay address (where payment is sent)
		payAddr, err := script.NewAddressFromString(listing.PayAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to create pay address: %w", err)
		}

		// Create P2PKH script for payment
		paymentScript, err := p2pkh.Lock(payAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create payment script: %w", err)
		}

		// Create the output for payment recipient
		payOutput := &transaction.TransactionOutput{
			LockingScript: paymentScript,
			Satoshis:      listing.Price,
		}

		// Create the OrdLock
		// Note: This is just placeholder code for future implementation
		_ = ordlock.OrdLock{
			Seller: sellerAddr,
			Price:  listing.Price,
			PayOut: payOutput.Bytes(),
		}

		// TODO: Create the actual OrdLock script
		// We need to use the go-templates ordlock package to create the script
		// This is a bit more complex than just creating a simple script

		// For now, we'll just use a P2PKH script as a placeholder
		// In a proper implementation, you'd use the ordlock template to create the script
		lockingScript, err := p2pkh.Lock(sellerAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create locking script: %w", err)
		}

		// Add the output to the transaction
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: lockingScript,
			Satoshis:      1, // 1 sat for ordinals
		})
	}

	// Add change output if needed
	if config.ChangeAddress != "" {
		changeAddr, err := script.NewAddressFromString(config.ChangeAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to create change address: %w", err)
		}

		changeScript, err := p2pkh.Lock(changeAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create change script: %w", err)
		}

		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: changeScript,
			Change:        true,
		})
	}

	// Set fee rate using SatsPerKb if provided, otherwise use the default value
	feeRate := uint64(DEFAULT_SAT_PER_KB)
	if config.SatsPerKb > 0 {
		feeRate = config.SatsPerKb
	}

	// Create fee model for computation
	feeModel := &fee_model.SatoshisPerKilobyte{
		Satoshis: feeRate,
	}

	err := tx.Fee(feeModel, transaction.ChangeDistributionEqual)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate fee: %w", err)
	}

	// Sign the transaction
	err = tx.Sign()
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}

// PurchaseOrdListing purchases an Ordinal Lock listing
func PurchaseOrdListing(config *PurchaseOrdListingConfig) (*transaction.Transaction, error) {
	// Create a new transaction
	tx := transaction.NewTransaction()

	// Add inputs
	for _, utxo := range config.Utxos {
		unlocker, err := p2pkh.Unlock(config.PaymentPk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create unlocker: %w", err)
		}

		err = tx.AddInputFrom(
			utxo.TxID,
			utxo.Vout,
			utxo.ScriptPubKey,
			utxo.Satoshis,
			unlocker,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add input: %w", err)
		}
	}

	// Add the ordinal listing input
	ordUtxo := config.ListingUtxo

	// TODO: Implement the proper unlocking for OrdLock
	// This would require implementing a custom unlocker
	// For now, we'll use a simple P2PKH unlocker as a placeholder
	unlocker, err := p2pkh.Unlock(config.PaymentPk, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create ordinal unlocker: %w", err)
	}

	err = tx.AddInputFrom(
		ordUtxo.TxID,
		ordUtxo.Vout,
		ordUtxo.ScriptPubKey,
		ordUtxo.Satoshis,
		unlocker,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add ordinal input: %w", err)
	}

	// Create output for the ordinal
	dstAddr, err := script.NewAddressFromString(config.OrdAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination address: %w", err)
	}

	lockingScript, err := p2pkh.Lock(dstAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create p2pkh script: %w", err)
	}

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: lockingScript,
		Satoshis:      1, // 1 sat for ordinals
	})

	// Add change output if needed
	if config.ChangeAddress != "" {
		changeAddr, err := script.NewAddressFromString(config.ChangeAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to create change address: %w", err)
		}

		changeScript, err := p2pkh.Lock(changeAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create change script: %w", err)
		}

		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: changeScript,
			Change:        true,
		})
	}

	// Set fee rate using SatsPerKb if provided, otherwise use the default value
	feeRate := uint64(DEFAULT_SAT_PER_KB)
	if config.SatsPerKb > 0 {
		feeRate = config.SatsPerKb
	}

	// Create fee model for computation
	feeModel := &fee_model.SatoshisPerKilobyte{
		Satoshis: feeRate,
	}

	err = tx.Fee(feeModel, transaction.ChangeDistributionEqual)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate fee: %w", err)
	}

	// Sign the transaction
	err = tx.Sign()
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}

// CancelOrdListings cancels an Ordinal Lock listing and returns the ordinal
func CancelOrdListings(config *CancelOrdListingsConfig) (*transaction.Transaction, error) {
	// Create a new transaction
	tx := transaction.NewTransaction()

	// Add inputs
	for _, utxo := range config.Utxos {
		unlocker, err := p2pkh.Unlock(config.PaymentPk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create unlocker: %w", err)
		}

		err = tx.AddInputFrom(
			utxo.TxID,
			utxo.Vout,
			utxo.ScriptPubKey,
			utxo.Satoshis,
			unlocker,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add input: %w", err)
		}
	}

	// Add the ordinal listing inputs
	for _, listingUtxo := range config.ListingUtxos {
		// TODO: Implement the proper unlocking for OrdLock (cancel path)
		// This would require implementing a custom unlocker
		// For now, we'll use a simple P2PKH unlocker as a placeholder
		unlocker, err := p2pkh.Unlock(config.OrdPk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create ordinal unlocker: %w", err)
		}

		err = tx.AddInputFrom(
			listingUtxo.TxID,
			listingUtxo.Vout,
			listingUtxo.ScriptPubKey,
			listingUtxo.Satoshis,
			unlocker,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add ordinal input: %w", err)
		}

		// Create output returning the ordinal to the original owner
		// Derive destination from OrdPk
		dstAddr, err := script.NewAddressFromPublicKey(config.OrdPk.PubKey(), true)
		if err != nil {
			return nil, fmt.Errorf("failed to create destination address: %w", err)
		}

		lockingScript, err := p2pkh.Lock(dstAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create p2pkh script: %w", err)
		}

		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: lockingScript,
			Satoshis:      1, // 1 sat for ordinals
		})
	}

	// Add change output if needed
	if config.ChangeAddress != "" {
		changeAddr, err := script.NewAddressFromString(config.ChangeAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to create change address: %w", err)
		}

		changeScript, err := p2pkh.Lock(changeAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create change script: %w", err)
		}

		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: changeScript,
			Change:        true,
		})
	}

	// Set fee rate using SatsPerKb if provided, otherwise use the default value
	feeRate := uint64(DEFAULT_SAT_PER_KB)
	if config.SatsPerKb > 0 {
		feeRate = config.SatsPerKb
	}

	// Create fee model for computation
	feeModel := &fee_model.SatoshisPerKilobyte{
		Satoshis: feeRate,
	}

	err := tx.Fee(feeModel, transaction.ChangeDistributionEqual)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate fee: %w", err)
	}

	// Sign the transaction
	err = tx.Sign()
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}

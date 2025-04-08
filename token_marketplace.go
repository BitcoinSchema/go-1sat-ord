package ordinals

import (
	"fmt"

	"github.com/bitcoin-sv/go-templates/template/bsv21"
	"github.com/bitcoin-sv/go-templates/template/ordlock"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	fee_model "github.com/bsv-blockchain/go-sdk/transaction/fee_model"
	"github.com/bsv-blockchain/go-sdk/transaction/template/p2pkh"
)

// PurchaseOrdTokenListing purchases a token listing
// It creates a transaction that:
// 1. Spends the locked token listing
// 2. Creates a new transfer inscription output
// 3. Makes the payment to the seller
// 4. Handles additional payments if specified
// 5. Calculates and includes the transaction fee
// 6. Returns change to the specified address
func PurchaseOrdTokenListing(config *PurchaseOrdTokenListingConfig) (*transaction.Transaction, error) {
	// Validate inputs
	if config.PaymentPk == nil {
		return nil, fmt.Errorf("private key is required to sign the transaction")
	}

	if config.ListingUtxo == nil {
		return nil, fmt.Errorf("listing UTXO is required")
	}

	if config.OrdAddress == "" {
		return nil, fmt.Errorf("destination address is required")
	}

	if config.ChangeAddress == "" && config.PaymentPk == nil {
		return nil, fmt.Errorf("either changeAddress or paymentPk is required")
	}

	// Create a new transaction
	tx := transaction.NewTransaction()

	// Add the locked token listing we're purchasing as an input
	listingUtxo := config.ListingUtxo

	// TODO: Once the ordlock implementation is complete, use proper unlocking
	// For now, we'll use a simple P2PKH unlocker as a placeholder
	unlocker, err := p2pkh.Unlock(config.PaymentPk, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create listing unlocker: %w", err)
	}

	err = tx.AddInputFrom(
		listingUtxo.TxID,
		listingUtxo.Vout,
		listingUtxo.ScriptPubKey,
		listingUtxo.Satoshis,
		unlocker,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add listing input: %w", err)
	}

	// Create output for the token transfer inscription
	dstAddr, err := script.NewAddressFromString(config.OrdAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination address: %w", err)
	}

	// Create token transfer data
	var transferData *bsv21.Bsv21
	if config.Protocol == TokenTypeBSV21 {
		transferData = &bsv21.Bsv21{
			Op:  string(bsv21.OpTransfer),
			Id:  config.TokenID,
			Amt: listingUtxo.Amount,
		}
	} else {
		return nil, fmt.Errorf("unsupported token protocol: %s", config.Protocol)
	}

	// Create P2PKH script for the destination
	p2pkhScript, err := p2pkh.Lock(dstAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create p2pkh script: %w", err)
	}

	// Create token script
	tokenScript, err := transferData.Lock(p2pkhScript)
	if err != nil {
		return nil, fmt.Errorf("failed to create token transfer script: %w", err)
	}

	// Add transfer output with the token
	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: tokenScript,
		Satoshis:      1, // 1 sat for ordinals
	})

	// Add payment output (to the seller)
	// TODO: In a real implementation, we would extract the payment details
	// from the listing's locking script or additional data
	// For now, we'll assume the payment is a simple P2PKH to the same address as the listing

	// Create payment script using the listing's original script
	// This is a placeholder - in a real implementation you'd decode the
	// payment data from the listing
	lockingScript, err := script.NewFromHex(listingUtxo.ScriptPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse seller script: %w", err)
	}

	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: lockingScript,
		Satoshis:      listingUtxo.Amount * 100, // Example price calculation
	})

	// Add additional payments if any
	if config.AdditionalPayments != nil {
		for _, payment := range config.AdditionalPayments {
			payAddr, err := script.NewAddressFromString(payment.Address)
			if err != nil {
				return nil, fmt.Errorf("failed to create payment address: %w", err)
			}

			payScript, err := p2pkh.Lock(payAddr)
			if err != nil {
				return nil, fmt.Errorf("failed to create payment script: %w", err)
			}

			tx.AddOutput(&transaction.TransactionOutput{
				LockingScript: payScript,
				Satoshis:      payment.Satoshis,
			})
		}
	}

	// Add payment inputs
	totalIn := uint64(0)
	for _, utxo := range config.Utxos {
		unlocker, err := p2pkh.Unlock(config.PaymentPk, nil)
		if err != nil {
			return nil, fmt.Errorf("private key is required to sign the payment: %w", err)
		}

		err = tx.AddInputFrom(
			utxo.TxID,
			utxo.Vout,
			utxo.ScriptPubKey,
			utxo.Satoshis,
			unlocker,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add payment input: %w", err)
		}

		totalIn += utxo.Satoshis
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
	feeRate := config.SatsPerKb
	if feeRate == 0 {
		feeRate = DEFAULT_SAT_PER_KB
	}

	// Create fee model for computation
	feeModel := &fee_model.SatoshisPerKilobyte{
		Satoshis: feeRate,
	}

	err = tx.Fee(feeModel, transaction.ChangeDistributionEqual)
	if err != nil {
		if err.Error() == "insufficient funds for fee" {
			return nil, fmt.Errorf("not enough funds to purchase token listing. Total sats in: %d", totalIn)
		}
		return nil, fmt.Errorf("failed to calculate fee: %w", err)
	}

	// Sign the transaction
	err = tx.Sign()
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}

// CreateOrdTokenListings creates token listings using an "Ordinal Lock" script
// It creates a transaction that:
// 1. Spends the token UTXOs
// 2. Creates new locked outputs with the tokens that can be purchased
// 3. Calculates and includes the transaction fee
// 4. Returns change to the specified address
func CreateOrdTokenListings(config *CreateOrdTokenListingsConfig) (*transaction.Transaction, error) {
	// Validate inputs
	if config.PaymentPk == nil {
		return nil, fmt.Errorf("payment private key is required to sign the transaction")
	}

	if config.OrdPk == nil {
		return nil, fmt.Errorf("token private key is required to sign the transaction")
	}

	if len(config.Listings) == 0 {
		return nil, fmt.Errorf("at least one listing is required")
	}

	if config.ChangeAddress == "" && config.PaymentPk == nil {
		return nil, fmt.Errorf("either changeAddress or paymentPk is required")
	}

	// Create a new transaction
	tx := transaction.NewTransaction()

	// Add payment inputs
	totalIn := uint64(0)
	for _, utxo := range config.Utxos {
		unlocker, err := p2pkh.Unlock(config.PaymentPk, nil)
		if err != nil {
			return nil, fmt.Errorf("private key is required to sign the payment: %w", err)
		}

		err = tx.AddInputFrom(
			utxo.TxID,
			utxo.Vout,
			utxo.ScriptPubKey,
			utxo.Satoshis,
			unlocker,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add payment input: %w", err)
		}

		totalIn += utxo.Satoshis
	}

	// Add token inputs and create locked outputs for each listing
	for _, listing := range config.Listings {
		// Validate listing data
		if listing.ListingUtxo == nil {
			return nil, fmt.Errorf("token UTXO is required for listing")
		}

		if listing.OrdAddress == "" {
			return nil, fmt.Errorf("token owner address is required for listing")
		}

		if listing.PayAddress == "" {
			return nil, fmt.Errorf("payment address is required for listing")
		}

		// Add the token input
		tokenUtxo := listing.ListingUtxo

		unlocker, err := p2pkh.Unlock(config.OrdPk, nil)
		if err != nil {
			return nil, fmt.Errorf("private key is required to sign the token input: %w", err)
		}

		err = tx.AddInputFrom(
			tokenUtxo.TxID,
			tokenUtxo.Vout,
			tokenUtxo.ScriptPubKey,
			tokenUtxo.Satoshis,
			unlocker,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add token input: %w", err)
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

		// Create the OrdLock for the token
		// Note: This is just placeholder code for future implementation
		_ = ordlock.OrdLock{
			Seller: sellerAddr,
			Price:  listing.Price,
			PayOut: payOutput.Bytes(),
		}

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
	feeRate := config.SatsPerKb
	if feeRate == 0 {
		feeRate = DEFAULT_SAT_PER_KB
	}

	// Create fee model for computation
	feeModel := &fee_model.SatoshisPerKilobyte{
		Satoshis: feeRate,
	}

	err := tx.Fee(feeModel, transaction.ChangeDistributionEqual)
	if err != nil {
		if err.Error() == "insufficient funds for fee" {
			return nil, fmt.Errorf("not enough funds to create token listings. Total sats in: %d", totalIn)
		}
		return nil, fmt.Errorf("failed to calculate fee: %w", err)
	}

	// Sign the transaction
	err = tx.Sign()
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}

// CancelOrdTokenListings cancels token listings and returns the tokens
// to the original owner.
// It creates a transaction that:
// 1. Spends the locked token listings
// 2. Creates new outputs for each token returned to the owner
// 3. Calculates and includes the transaction fee
// 4. Returns change to the specified address
func CancelOrdTokenListings(config *CancelOrdTokenListingsConfig) (*transaction.Transaction, error) {
	// Validate inputs
	if config.PaymentPk == nil {
		return nil, fmt.Errorf("payment private key is required to sign the transaction")
	}

	if config.OrdPk == nil {
		return nil, fmt.Errorf("token private key is required to sign the transaction")
	}

	if len(config.ListingUtxos) == 0 {
		return nil, fmt.Errorf("at least one listing UTXO is required")
	}

	if config.ChangeAddress == "" && config.PaymentPk == nil {
		return nil, fmt.Errorf("either changeAddress or paymentPk is required")
	}

	// Create a new transaction
	tx := transaction.NewTransaction()

	// Add payment inputs (for fees)
	totalIn := uint64(0)
	for _, utxo := range config.Utxos {
		unlocker, err := p2pkh.Unlock(config.PaymentPk, nil)
		if err != nil {
			return nil, fmt.Errorf("private key is required to sign the payment: %w", err)
		}

		err = tx.AddInputFrom(
			utxo.TxID,
			utxo.Vout,
			utxo.ScriptPubKey,
			utxo.Satoshis,
			unlocker,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add payment input: %w", err)
		}

		totalIn += utxo.Satoshis
	}

	// Add listing inputs and create outputs for each token
	for _, listingUtxo := range config.ListingUtxos {
		// Validate listing UTXO
		if listingUtxo.Protocol == "" {
			return nil, fmt.Errorf("token protocol is required for listing UTXO")
		}

		if listingUtxo.TokenID == "" {
			return nil, fmt.Errorf("token ID is required for listing UTXO")
		}

		// TODO: In a real implementation, we would use OrdLock.Unlock
		// For now, we'll use a placeholder P2PKH unlocker
		unlocker, err := p2pkh.Unlock(config.OrdPk, nil)
		if err != nil {
			return nil, fmt.Errorf("private key is required to sign the listing: %w", err)
		}

		err = tx.AddInputFrom(
			listingUtxo.TxID,
			listingUtxo.Vout,
			listingUtxo.ScriptPubKey,
			listingUtxo.Satoshis,
			unlocker,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add listing input: %w", err)
		}

		// Create address for the token to be returned to
		// In a real implementation, we would extract this from the OrdLock script
		// For now, we'll assume the token should go back to the same address that signed the input
		tokenAddress, err := script.NewAddressFromPublicKey(config.OrdPk.PubKey(), true)
		if err != nil {
			return nil, fmt.Errorf("failed to create token address: %w", err)
		}

		// Create token transfer data
		var transferData *bsv21.Bsv21
		if listingUtxo.Protocol == TokenTypeBSV21 {
			transferData = &bsv21.Bsv21{
				Op:  string(bsv21.OpTransfer),
				Id:  listingUtxo.TokenID,
				Amt: listingUtxo.Amount,
			}
		} else {
			return nil, fmt.Errorf("unsupported token protocol: %s", listingUtxo.Protocol)
		}

		// Create P2PKH script for the destination
		p2pkhScript, err := p2pkh.Lock(tokenAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to create p2pkh script: %w", err)
		}

		// Create token script
		tokenScript, err := transferData.Lock(p2pkhScript)
		if err != nil {
			return nil, fmt.Errorf("failed to create token transfer script: %w", err)
		}

		// Add output for the token
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: tokenScript,
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
	feeRate := config.SatsPerKb
	if feeRate == 0 {
		feeRate = DEFAULT_SAT_PER_KB
	}

	// Create fee model for computation
	feeModel := &fee_model.SatoshisPerKilobyte{
		Satoshis: feeRate,
	}

	err := tx.Fee(feeModel, transaction.ChangeDistributionEqual)
	if err != nil {
		if err.Error() == "insufficient funds for fee" {
			return nil, fmt.Errorf("not enough funds to cancel token listings. Total sats in: %d", totalIn)
		}
		return nil, fmt.Errorf("failed to calculate fee: %w", err)
	}

	// Sign the transaction
	err = tx.Sign()
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}

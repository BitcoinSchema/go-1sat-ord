package ordinals

import (
	"fmt"

	"github.com/bitcoin-sv/go-templates/template/ordp2pkh"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	feemodel "github.com/bsv-blockchain/go-sdk/transaction/fee_model"
	"github.com/bsv-blockchain/go-sdk/transaction/template/p2pkh"
)

// CreateOrdinals creates a transaction with inscription outputs
func CreateOrdinals(config *CreateOrdinalsConfig) (*transaction.Transaction, error) {
	// Create a new transaction
	tx := transaction.NewTransaction()

	// Warn if creating many inscriptions at once (log a warning message)
	if len(config.Destinations) > 100 {
		// In Go we could use the log package, but for consistency we'll just print
		fmt.Println("WARNING: Creating many inscriptions at once can be slow. Consider using multiple transactions instead.")
	}

	// Add inputs
	for _, utxo := range config.Utxos {
		unlocker, err := p2pkh.Unlock(config.PaymentPk, nil)
		if err != nil {
			return nil, fmt.Errorf("private key is required to sign the transaction: %w", err)
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

	// Add ordinal inscription outputs
	for _, dest := range config.Destinations {
		// Validate destination has necessary data
		if dest.Inscription == nil {
			return nil, fmt.Errorf("inscription is required for all destinations")
		}

		// Create the destination address
		dstAddr, err := script.NewAddressFromString(dest.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to create destination address: %w", err)
		}

		var lockingScript *script.Script

		if dest.OmitMetadata() {
			// If omitMetadata is enabled, use a simple P2PKH output
			lockingScript, err = p2pkh.Lock(dstAddr)
			if err != nil {
				return nil, fmt.Errorf("failed to create p2pkh script: %w", err)
			}
		} else {
			// Create the ordinal P2PKH script with the inscription
			ordP2pkh := &ordp2pkh.OrdP2PKH{
				Inscription: dest.Inscription,
				Address:     dstAddr,
			}

			// Get the locking script
			lockingScript, err = ordP2pkh.Lock()
			if err != nil {
				return nil, fmt.Errorf("failed to create ordp2pkh locking script: %w", err)
			}
		}

		// Add the output to the transaction
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: lockingScript,
			Satoshis:      1, // 1 sat for ordinals
		})
	}

	// Add additional payments if provided
	if config.AdditionalPayments != nil {
		for _, payment := range config.AdditionalPayments {
			dstAddr, err := script.NewAddressFromString(payment.Address)
			if err != nil {
				return nil, fmt.Errorf("failed to create payment address: %w", err)
			}

			lockingScript, err := p2pkh.Lock(dstAddr)
			if err != nil {
				return nil, fmt.Errorf("failed to create payment script: %w", err)
			}

			tx.AddOutput(&transaction.TransactionOutput{
				LockingScript: lockingScript,
				Satoshis:      payment.Satoshis,
			})
		}
	}

	// Add change output if needed
	if config.ChangeAddress == "" && config.PaymentPk == nil {
		return nil, fmt.Errorf("either changeAddress or paymentPk is required")
	}

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

	// Set fee rate using SatsPerKb if provided, otherwise use the default value
	feeRate := config.SatsPerKb
	if feeRate == 0 {
		feeRate = DEFAULT_SAT_PER_KB
	}

	// Create fee model for computation
	feeModel := &feemodel.SatoshisPerKilobyte{
		Satoshis: feeRate,
	}

	// Calculate total inputs and outputs for funds check
	totalIn := uint64(0)
	for _, utxo := range config.Utxos {
		totalIn += utxo.Satoshis
	}

	// Calculate and set fee
	err = tx.Fee(feeModel, transaction.ChangeDistributionEqual)
	if err != nil {
		if err.Error() == "insufficient funds for fee" {
			return nil, fmt.Errorf("not enough funds to create ordinals. Total sats in: %d", totalIn)
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

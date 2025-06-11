package ordinals

import (
	"fmt"

	"github.com/bitcoin-sv/go-templates/template/ordp2pkh"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	fee_model "github.com/bsv-blockchain/go-sdk/transaction/fee_model"
	"github.com/bsv-blockchain/go-sdk/transaction/template/p2pkh"
)

// SendOrdinals sends ordinals to the given destinations
func SendOrdinals(config *SendOrdinalsConfig) (*transaction.Transaction, error) {
	// Create a new transaction
	tx := transaction.NewTransaction()

	// Set defaults for optional parameters
	feeRate := config.SatsPerKb
	if feeRate == 0 {
		feeRate = DEFAULT_SAT_PER_KB
	}

	// Set a default for enforceUniformSend if it's not provided
	enforceUniform := config.EnforceUniformSend

	// If enforceUniformSend is true, check that the number of destinations matches the number of ordinals
	if enforceUniform && (len(config.Destinations) != len(config.Ordinals)) {
		return nil, fmt.Errorf("number of destinations must match number of ordinals being sent")
	}

	// Add ordinal inputs first
	for _, ordinalUtxo := range config.Ordinals {
		// Verify that ordinals have exactly 1 satoshi
		if ordinalUtxo.Satoshis != 1 {
			return nil, fmt.Errorf("1Sat Ordinal utxos must have exactly 1 satoshi")
		}

		unlocker, err := p2pkh.Unlock(config.OrdPk, nil)
		if err != nil {
			return nil, fmt.Errorf("private key is required to sign the ordinal: %w", err)
		}

		err = tx.AddInputFrom(
			ordinalUtxo.TxID,
			ordinalUtxo.Vout,
			ordinalUtxo.ScriptPubKey,
			ordinalUtxo.Satoshis,
			unlocker,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to add ordinal input: %w", err)
		}
	}

	// Add payment inputs
	for _, utxo := range config.PaymentUtxos {
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
	}

	// Add outputs for each destination
	for _, dest := range config.Destinations {
		// Create the destination address
		dstAddr, err := script.NewAddressFromString(dest.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to create destination address: %w", err)
		}

		var lockingScript *script.Script

		// Check if we should omit metadata
		if dest.OmitMetadata() {
			// If omitMetadata is enabled, use a simple P2PKH output
			lockingScript, err = p2pkh.Lock(dstAddr)
			if err != nil {
				return nil, fmt.Errorf("failed to create p2pkh script: %w", err)
			}
		} else if dest.Inscription != nil {
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
		} else {
			// Just create a regular P2PKH output
			lockingScript, err = p2pkh.Lock(dstAddr)
			if err != nil {
				return nil, fmt.Errorf("failed to create p2pkh script: %w", err)
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

	// Calculate total inputs and outputs for funds check
	totalIn := uint64(0)
	for _, utxo := range config.PaymentUtxos {
		totalIn += utxo.Satoshis
	}

	// Create fee model for computation
	feeModel := &fee_model.SatoshisPerKilobyte{
		Satoshis: feeRate,
	}

	err := tx.Fee(feeModel, transaction.ChangeDistributionEqual)
	if err != nil {
		if err.Error() == "insufficient funds for fee" {
			return nil, fmt.Errorf("not enough funds to send ordinals. Total sats in: %d", totalIn)
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

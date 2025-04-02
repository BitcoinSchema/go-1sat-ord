package ordinals

import (
	"fmt"

	"github.com/bitcoin-sv/go-templates/template/inscription"
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

	// Add ordinal inscription outputs
	for _, dest := range config.Destinations {
		// Create the destination address
		dstAddr, err := script.NewAddressFromString(dest.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to create destination address: %w", err)
		}

		// Create the inscription
		insc := &inscription.Inscription{
			File: inscription.File{
				Content: dest.File.Content,
				Type:    dest.File.ContentType,
			},
		}

		// Add metadata if present
		if len(dest.Metadata) > 0 {
			insc.Bitcom = dest.Metadata
		}

		// Create the ordinal P2PKH script
		ordP2pkh := &ordp2pkh.OrdP2PKH{
			Inscription: insc,
			Address:     dstAddr,
		}

		// Get the locking script
		lockingScript, err := ordP2pkh.Lock()
		if err != nil {
			return nil, fmt.Errorf("failed to create ordp2pkh locking script: %w", err)
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
	feeModel := &feemodel.SatoshisPerKilobyte{
		Satoshis: feeRate,
	}

	// Calculate and set fee
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

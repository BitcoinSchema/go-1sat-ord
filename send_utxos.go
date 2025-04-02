package ordinals

import (
	"fmt"

	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	fee_model "github.com/bsv-blockchain/go-sdk/transaction/fee_model"
	"github.com/bsv-blockchain/go-sdk/transaction/template/p2pkh"
)

// SendUtxos sends utxos to the given destinations
func SendUtxos(config *SendUtxosConfig) (*transaction.Transaction, error) {
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

	// Add payment outputs
	for _, payment := range config.Payments {
		dstAddr, err := script.NewAddressFromString(payment.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to create destination address: %w", err)
		}

		lockingScript, err := p2pkh.Lock(dstAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create p2pkh script: %w", err)
		}

		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: lockingScript,
			Satoshis:      payment.Satoshis,
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
		return nil, fmt.Errorf("failed to calculate fee: %w", err)
	}

	// Sign the transaction
	err = tx.Sign()
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}

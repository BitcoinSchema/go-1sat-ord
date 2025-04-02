package ordinals

import (
	"encoding/hex"
	"fmt"
	"strings"

	ec "github.com/bsv-blockchain/go-sdk/primitives/ec"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	feemodel "github.com/bsv-blockchain/go-sdk/transaction/fee_model"
	"github.com/bsv-blockchain/go-sdk/transaction/template/p2pkh"
)

// BurnOrdinalsConfig represents configuration for burning ordinals
type BurnOrdinalsConfig struct {
	// PaymentUtxos is the list of UTXOs to use for paying the fee
	PaymentUtxos []*Utxo
	// PaymentPk is the private key for the payment UTXOs
	PaymentPk *ec.PrivateKey
	// Ordinals is the list of NFT UTXOs to burn
	Ordinals []*NftUtxo
	// OrdPk is the private key for the ordinals
	OrdPk *ec.PrivateKey
	// Metadata is optional MAP protocol metadata to include in an OP_RETURN output
	Metadata map[string][]byte
	// ChangeAddress is the address to send any change to
	ChangeAddress string
	// SatsPerKb is the fee rate in satoshis per kilobyte
	SatsPerKb uint64
}

// BurnOrdinals burns ordinals by consuming them as fees
// It creates a transaction that spends the ordinal UTXOs and adds an optional
// OP_RETURN output with MAP protocol metadata
func BurnOrdinals(config *BurnOrdinalsConfig) (*transaction.Transaction, error) {
	// Create a new transaction
	tx := transaction.NewTransaction()

	// Add payment inputs
	for _, utxo := range config.PaymentUtxos {
		unlocker, err := p2pkh.Unlock(config.PaymentPk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create payment unlocker: %w", err)
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

	// Add ordinal inputs
	for _, ordUtxo := range config.Ordinals {
		// Check that it's a 1-sat ordinal
		if ordUtxo.Satoshis != 1 {
			return nil, fmt.Errorf("1Sat Ordinal UTXOs must have exactly 1 satoshi")
		}

		// Create the unlocker
		unlocker, err := p2pkh.Unlock(config.OrdPk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create ordinal unlocker: %w", err)
		}

		// Add the input
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
	}

	// Add OP_RETURN output with metadata if provided
	if len(config.Metadata) > 0 {
		// Build a simple OP_RETURN script with MAP metadata
		asm := "OP_FALSE OP_RETURN "

		// Add MAP prefix
		mapPrefix := []byte(MAP_PREFIX)
		asm += hex.EncodeToString(mapPrefix) + " "

		// Add SET command
		asm += hex.EncodeToString([]byte("SET")) + " "

		// Add metadata entries
		for key, value := range config.Metadata {
			asm += hex.EncodeToString([]byte(key)) + " "
			asm += hex.EncodeToString(value) + " "
		}

		// Remove trailing space
		asm = strings.TrimSpace(asm)

		// Create the script
		opReturnScript, err := script.NewFromASM(asm)
		if err != nil {
			return nil, fmt.Errorf("failed to create OP_RETURN script: %w", err)
		}

		// Add output with the inscription script
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: opReturnScript,
			Satoshis:      0, // 0 sats for OP_RETURN
		})
	} else {
		// Create a simple OP_FALSE OP_RETURN output
		scriptAsm, err := script.NewFromASM("OP_FALSE OP_RETURN")
		if err != nil {
			return nil, fmt.Errorf("failed to create OP_RETURN script: %w", err)
		}

		// Add output with the simple OP_RETURN script
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: scriptAsm,
			Satoshis:      0,
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

	// Set the fee rate
	feeRate := DEFAULT_SAT_PER_KB
	if config.SatsPerKb > 0 {
		feeRate = config.SatsPerKb
	}

	// Create fee model
	feeModel := &feemodel.SatoshisPerKilobyte{
		Satoshis: feeRate,
	}

	// Calculate fee
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

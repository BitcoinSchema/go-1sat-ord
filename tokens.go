package ordinals

import (
	"fmt"

	"github.com/bitcoin-sv/go-templates/template/bsv21"
	"github.com/bsv-blockchain/go-sdk/script"
	"github.com/bsv-blockchain/go-sdk/transaction"
	fee_model "github.com/bsv-blockchain/go-sdk/transaction/fee_model"
	"github.com/bsv-blockchain/go-sdk/transaction/template/p2pkh"
)

// DeployBsv21Token deploys a new BSV21 token
func DeployBsv21Token(config *DeployBsv21TokenConfig) (*transaction.Transaction, error) {
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

	// Create the destination address for the token
	dstAddr, err := script.NewAddressFromString(config.DestinationAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination address: %w", err)
	}

	// Create the P2PKH script for the destination
	p2pkhScript, err := p2pkh.Lock(dstAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create p2pkh script: %w", err)
	}

	// Create the BSV21 token
	decimals := uint8(0)
	token := &bsv21.Bsv21{
		Op:       string(bsv21.OpMint),
		Symbol:   &config.Symbol,
		Decimals: &decimals,
		Amt:      uint64(config.InitialDistribution.Tokens),
	}

	// Add icon if specified
	if config.Icon != "" {
		token.Icon = &config.Icon
	}

	// Create the token script
	tokenScript, err := token.Lock(p2pkhScript)
	if err != nil {
		return nil, fmt.Errorf("failed to create token script: %w", err)
	}

	// Add the token output to the transaction
	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: tokenScript,
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
		return nil, fmt.Errorf("failed to calculate fee: %w", err)
	}

	// Sign the transaction
	err = tx.Sign()
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return tx, nil
}

// TransferOrdToken transfers BSV21 tokens
func TransferOrdToken(config *TransferBsv21TokenConfig) (*transaction.Transaction, error) {
	// Check protocol type
	if config.Protocol != TokenTypeBSV21 {
		return nil, fmt.Errorf("unsupported token protocol: %s", config.Protocol)
	}

	// Create a new transaction
	tx := transaction.NewTransaction()

	// Add payment inputs
	for _, utxo := range config.Utxos {
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

	// Add token inputs
	var totalTokens uint64
	for _, tokenUtxo := range config.InputTokens {
		unlocker, err := p2pkh.Unlock(config.OrdPk, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create token unlocker: %w", err)
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

		totalTokens += tokenUtxo.Amount
	}

	// Process token distributions
	var distributedTokens uint64
	for _, dist := range config.Distributions {
		// Calculate token amount
		tokenAmount := uint64(dist.Tokens)
		distributedTokens += tokenAmount

		// Create destination address
		dstAddr, err := script.NewAddressFromString(dist.Address)
		if err != nil {
			return nil, fmt.Errorf("failed to create destination address: %w", err)
		}

		// Create P2PKH script
		p2pkhScript, err := p2pkh.Lock(dstAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to create p2pkh script: %w", err)
		}

		// Create token transfer
		token := &bsv21.Bsv21{
			Op:  string(bsv21.OpTransfer),
			Id:  config.TokenID,
			Amt: tokenAmount,
		}

		// Create token script
		// Note: The OmitMetadata flag is currently not used as the bsv21 library
		// doesn't support creating scripts without the token data.
		// Future enhancement: When the library supports it, this would create
		// a minimal script when dist.OmitMetadata is true.
		tokenScript, err := token.Lock(p2pkhScript)
		if err != nil {
			return nil, fmt.Errorf("failed to create token script: %w", err)
		}

		// Add token output
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: tokenScript,
			Satoshis:      1, // 1 sat for ordinals
		})
	}

	// Handle remaining tokens as change
	remainingTokens := totalTokens - distributedTokens

	// Only process change if not burning tokens and there are tokens remaining
	if remainingTokens > 0 && !config.Burn {
		var err error
		// Check token input mode - if "needed", we might not need to consume all inputs
		if config.TokenInputMode == TokenInputModeNeeded && distributedTokens < totalTokens {
			// Create split outputs if configured
			if config.SplitConfig != nil && config.SplitConfig.Outputs > 1 {
				// Add split token outputs
				err = createSplitTokenOutputs(tx, config, remainingTokens)
				if err != nil {
					return nil, err
				}
			} else {
				// Create a single change output
				err = createSingleTokenChangeOutput(tx, config, remainingTokens)
				if err != nil {
					return nil, err
				}
			}
		} else if config.TokenInputMode == TokenInputModeAll || config.TokenInputMode == "" {
			// Default is to use all inputs
			// Create split outputs if configured
			if config.SplitConfig != nil && config.SplitConfig.Outputs > 1 {
				// Add split token outputs
				err = createSplitTokenOutputs(tx, config, remainingTokens)
				if err != nil {
					return nil, err
				}
			} else {
				// Create a single change output
				err = createSingleTokenChangeOutput(tx, config, remainingTokens)
				if err != nil {
					return nil, err
				}
			}
		}
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

// createSplitTokenOutputs splits token change into multiple outputs according to config
// Note: SplitConfig.OmitMetadata is currently just a flag for future enhancement.
// Current implementation always includes the token data in the script.
func createSplitTokenOutputs(
	tx *transaction.Transaction,
	config *TransferBsv21TokenConfig,
	remainingTokens uint64,
) error {
	outputs := config.SplitConfig.Outputs

	// Default threshold is 0 if not specified
	var threshold uint64 = 0
	if config.SplitConfig.Threshold != nil {
		threshold = uint64(*config.SplitConfig.Threshold)
	}

	// Calculate tokens per split
	tokensPerSplit := remainingTokens / uint64(outputs)

	// If tokens per split is below threshold, reduce the number of outputs
	// This ensures each output has at least the threshold amount of tokens
	if tokensPerSplit < threshold && threshold > 0 {
		outputs = int(remainingTokens / threshold)
		if outputs == 0 {
			outputs = 1 // Ensure at least one output
		}
		tokensPerSplit = remainingTokens / uint64(outputs)
	}

	// Get address for token change (use OrdPk to derive address)
	dstAddr, err := script.NewAddressFromPublicKey(config.OrdPk.PubKey(), true)
	if err != nil {
		return fmt.Errorf("failed to create token change address: %w", err)
	}

	// Create P2PKH script
	p2pkhScript, err := p2pkh.Lock(dstAddr)
	if err != nil {
		return fmt.Errorf("failed to create p2pkh script: %w", err)
	}

	// Distribute tokens across outputs
	tokensLeft := remainingTokens
	for i := 0; i < outputs && tokensLeft > 0; i++ {
		// Last output gets any remainder
		outputAmount := tokensPerSplit
		if i == outputs-1 || tokensLeft < tokensPerSplit {
			outputAmount = tokensLeft
		}

		// Create token transfer
		token := &bsv21.Bsv21{
			Op:  string(bsv21.OpTransfer),
			Id:  config.TokenID,
			Amt: outputAmount,
		}

		// Create token script
		// Note: The OmitMetadata flag is just for future enhancement.
		// Currently, all token scripts include the token data as an inscription.
		tokenScript, err := token.Lock(p2pkhScript)
		if err != nil {
			return fmt.Errorf("failed to create token script: %w", err)
		}

		// Add token change output
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: tokenScript,
			Satoshis:      1, // 1 sat for ordinals
		})

		tokensLeft -= outputAmount
	}

	return nil
}

// createSingleTokenChangeOutput creates a single token change output
// Note: SplitConfig.OmitMetadata is currently just a flag for future enhancement.
// Current implementation always includes the token data in the script.
func createSingleTokenChangeOutput(
	tx *transaction.Transaction,
	config *TransferBsv21TokenConfig,
	remainingTokens uint64,
) error {
	// Get address for token change (use OrdPk to derive address)
	dstAddr, err := script.NewAddressFromPublicKey(config.OrdPk.PubKey(), true)
	if err != nil {
		return fmt.Errorf("failed to create token change address: %w", err)
	}

	// Create P2PKH script
	p2pkhScript, err := p2pkh.Lock(dstAddr)
	if err != nil {
		return fmt.Errorf("failed to create p2pkh script: %w", err)
	}

	// Create token change
	token := &bsv21.Bsv21{
		Op:  string(bsv21.OpTransfer),
		Id:  config.TokenID,
		Amt: remainingTokens,
	}

	// Create token script
	// Note: The OmitMetadata flag is just for future enhancement.
	// Currently, all token scripts include the token data as an inscription.
	tokenScript, err := token.Lock(p2pkhScript)
	if err != nil {
		return fmt.Errorf("failed to create token script: %w", err)
	}

	// Add token change output
	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: tokenScript,
		Satoshis:      1, // 1 sat for ordinals
	})

	return nil
}

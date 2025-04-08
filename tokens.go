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
	// Validate input params
	if config.Symbol == "" {
		return nil, fmt.Errorf("token symbol is required")
	}

	if config.InitialDistribution == nil {
		return nil, fmt.Errorf("initial distribution is required")
	}

	if config.InitialDistribution.Tokens <= 0 {
		return nil, fmt.Errorf("initial distribution amount must be greater than zero")
	}

	// Create a new transaction
	tx := transaction.NewTransaction()

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

	// Ensure we have a change address
	if config.ChangeAddress == "" && config.PaymentPk == nil {
		return nil, fmt.Errorf("either changeAddress or paymentPk is required")
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

	// Calculate total inputs for funds checking
	totalIn := uint64(0)
	for _, utxo := range config.Utxos {
		totalIn += utxo.Satoshis
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
			return nil, fmt.Errorf("not enough funds to deploy token. Total sats in: %d", totalIn)
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

// TransferOrdTokens transfers BSV21 tokens
// This function is renamed to match the TypeScript version (transferOrdTokens)
func TransferOrdTokens(config *TransferBsv21TokenConfig) (*transaction.Transaction, error) {
	// Check protocol type
	if config.Protocol != TokenTypeBSV21 {
		return nil, fmt.Errorf("invalid protocol: expected %s, got %s", TokenTypeBSV21, config.Protocol)
	}

	// Ensure input tokens match the expected tokenID
	for _, token := range config.InputTokens {
		if token.TokenID != config.TokenID {
			return nil, fmt.Errorf("input tokens do not match the provided tokenID")
		}
	}

	// Create a new transaction
	tx := transaction.NewTransaction()

	// Add token inputs
	var totalTokens uint64
	for _, tokenUtxo := range config.InputTokens {
		unlocker, err := p2pkh.Unlock(config.OrdPk, nil)
		if err != nil {
			return nil, fmt.Errorf("private key required for token input: %w", err)
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

		var lockingScript *script.Script

		if dist.OmitMetadata {
			// If OmitMetadata is enabled, use the P2PKH script directly without token data
			lockingScript = p2pkhScript
		} else {
			// Create token transfer with metadata (normal case)
			token := &bsv21.Bsv21{
				Op:  string(bsv21.OpTransfer),
				Id:  config.TokenID,
				Amt: tokenAmount,
			}

			// Create token script
			var err error
			lockingScript, err = token.Lock(p2pkhScript)
			if err != nil {
				return nil, fmt.Errorf("failed to create token transfer script: %w", err)
			}
		}

		// Add the token output to the transaction
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: lockingScript,
			Satoshis:      1, // 1 sat for ordinals
		})
	}

	// Check if we have enough tokens
	if distributedTokens > totalTokens {
		return nil, fmt.Errorf("not enough tokens to satisfy the transfer amount")
	}

	// Handle remaining tokens if not burning
	remainingTokens := totalTokens - distributedTokens
	if remainingTokens > 0 && !config.Burn {
		// Ensure we have a change address
		if config.ChangeAddress == "" && config.PaymentPk == nil {
			return nil, fmt.Errorf("ordPk or changeAddress required for token change")
		}

		// Handle token change outputs based on input mode and split config
		if config.TokenInputMode == TokenInputModeAll || config.TokenInputMode == "" {
			// If in "all" mode or not specified, we must handle all change
			if config.SplitConfig != nil && config.SplitConfig.Outputs > 1 {
				err := createSplitTokenOutputs(tx, config, remainingTokens)
				if err != nil {
					return nil, err
				}
			} else {
				err := createSingleTokenChangeOutput(tx, config, remainingTokens)
				if err != nil {
					return nil, err
				}
			}
		} else if config.TokenInputMode == TokenInputModeNeeded {
			// In "needed" mode, we only use what's required, so add a single change output
			err := createSingleTokenChangeOutput(tx, config, remainingTokens)
			if err != nil {
				return nil, err
			}
		}
	}

	// Add payment inputs
	for _, utxo := range config.Utxos {
		unlocker, err := p2pkh.Unlock(config.PaymentPk, nil)
		if err != nil {
			return nil, fmt.Errorf("private key required for payment utxo: %w", err)
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

	// Calculate total inputs for funds checking
	totalIn := uint64(0)
	for _, utxo := range config.Utxos {
		totalIn += utxo.Satoshis
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
			return nil, fmt.Errorf("not enough funds to transfer tokens. Total sats in: %d", totalIn)
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

// createSplitTokenOutputs splits token change into multiple outputs according to config
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

		var lockingScript *script.Script

		if config.SplitConfig.OmitMetadata {
			// If OmitMetadata is enabled, use the P2PKH script directly without token data
			lockingScript = p2pkhScript
		} else {
			// Create token transfer with metadata (normal case)
			token := &bsv21.Bsv21{
				Op:  string(bsv21.OpTransfer),
				Id:  config.TokenID,
				Amt: outputAmount,
			}

			// Create token script with full metadata
			var err error
			lockingScript, err = token.Lock(p2pkhScript)
			if err != nil {
				return fmt.Errorf("failed to create token script: %w", err)
			}
		}

		// Add token change output
		tx.AddOutput(&transaction.TransactionOutput{
			LockingScript: lockingScript,
			Satoshis:      1, // 1 sat for ordinals
		})

		tokensLeft -= outputAmount
	}

	return nil
}

// createSingleTokenChangeOutput creates a single token change output
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

	var lockingScript *script.Script

	// Check if we should omit metadata (when using SplitConfig but just creating 1 output)
	omitMetadata := config.SplitConfig != nil && config.SplitConfig.OmitMetadata

	if omitMetadata {
		// If OmitMetadata is enabled, use the P2PKH script directly without token data
		lockingScript = p2pkhScript
	} else {
		// Create token change with metadata (normal case)
		token := &bsv21.Bsv21{
			Op:  string(bsv21.OpTransfer),
			Id:  config.TokenID,
			Amt: remainingTokens,
		}

		// Create token script with full metadata
		var err error
		lockingScript, err = token.Lock(p2pkhScript)
		if err != nil {
			return fmt.Errorf("failed to create token script: %w", err)
		}
	}

	// Add token change output
	tx.AddOutput(&transaction.TransactionOutput{
		LockingScript: lockingScript,
		Satoshis:      1, // 1 sat for ordinals
	})

	return nil
}

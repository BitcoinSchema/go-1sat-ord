package ordinals

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectTokenUtxos(t *testing.T) {
	// Create mock token UTXOs for testing
	mockUtxos := []*TokenUtxo{
		{
			Utxo: Utxo{
				TxID:         "tx1",
				Vout:         0,
				ScriptPubKey: "script1",
				Satoshis:     1,
			},
			TokenID:  "token1",
			Protocol: TokenTypeBSV21,
			Amount:   100,
			Decimals: 2,
		},
		{
			Utxo: Utxo{
				TxID:         "tx2",
				Vout:         1,
				ScriptPubKey: "script2",
				Satoshis:     1,
			},
			TokenID:  "token1",
			Protocol: TokenTypeBSV21,
			Amount:   200,
			Decimals: 2,
		},
		{
			Utxo: Utxo{
				TxID:         "tx3",
				Vout:         2,
				ScriptPubKey: "script3",
				Satoshis:     1,
			},
			TokenID:  "token1",
			Protocol: TokenTypeBSV21,
			Amount:   300,
			Decimals: 2,
		},
		{
			Utxo: Utxo{
				TxID:         "tx4",
				Vout:         3,
				ScriptPubKey: "script4",
				Satoshis:     1,
			},
			TokenID:  "token1",
			Protocol: TokenTypeBSV21,
			Amount:   400,
			Decimals: 2,
		},
		{
			Utxo: Utxo{
				TxID:         "tx5",
				Vout:         4,
				ScriptPubKey: "script5",
				Satoshis:     1,
			},
			TokenID:  "token1",
			Protocol: TokenTypeBSV21,
			Amount:   500,
			Decimals: 2,
		},
	}

	// Test case 1: Default strategy (RetainOrder for input and output)
	t.Run("DefaultStrategy", func(t *testing.T) {
		result := SelectTokenUtxos(mockUtxos, 5.5, 2, nil)
		assert.Equal(t, mockUtxos[:3], result.SelectedUtxos)
		assert.Equal(t, 6.0, result.TotalSelected)
		assert.True(t, result.IsEnough)
	})

	// Test case 2: SmallestFirst output strategy
	t.Run("SmallestFirstOutputStrategy", func(t *testing.T) {
		options := &TokenSelectionOptions{
			OutputStrategy: TokenSelectionStrategySmallestFirst,
		}
		result := SelectTokenUtxos(mockUtxos, 10.0, 2, options)
		assert.Equal(t, 4, len(result.SelectedUtxos))
		assert.Equal(t, mockUtxos[0].Amount, result.SelectedUtxos[0].Amount)
		assert.Equal(t, mockUtxos[1].Amount, result.SelectedUtxos[1].Amount)
		assert.Equal(t, mockUtxos[2].Amount, result.SelectedUtxos[2].Amount)
		assert.Equal(t, mockUtxos[3].Amount, result.SelectedUtxos[3].Amount)
		assert.Equal(t, 10.0, result.TotalSelected)
		assert.True(t, result.IsEnough)
	})

	// Test case 3: LargestFirst output strategy
	t.Run("LargestFirstOutputStrategy", func(t *testing.T) {
		options := &TokenSelectionOptions{
			OutputStrategy: TokenSelectionStrategyLargestFirst,
		}
		result := SelectTokenUtxos(mockUtxos, 10.0, 2, options)
		assert.Equal(t, 4, len(result.SelectedUtxos))
		assert.Equal(t, mockUtxos[3].Amount, result.SelectedUtxos[0].Amount)
		assert.Equal(t, mockUtxos[2].Amount, result.SelectedUtxos[1].Amount)
		assert.Equal(t, mockUtxos[1].Amount, result.SelectedUtxos[2].Amount)
		assert.Equal(t, mockUtxos[0].Amount, result.SelectedUtxos[3].Amount)
		assert.Equal(t, 10.0, result.TotalSelected)
		assert.True(t, result.IsEnough)
	})

	// Test case 4: LargestFirst input strategy
	t.Run("LargestFirstInputStrategy", func(t *testing.T) {
		options := &TokenSelectionOptions{
			InputStrategy: TokenSelectionStrategyLargestFirst,
		}
		result := SelectTokenUtxos(mockUtxos, 10.0, 2, options)
		assert.Equal(t, 3, len(result.SelectedUtxos))
		assert.Equal(t, mockUtxos[4].Amount, result.SelectedUtxos[0].Amount)
		assert.Equal(t, mockUtxos[3].Amount, result.SelectedUtxos[1].Amount)
		assert.Equal(t, mockUtxos[2].Amount, result.SelectedUtxos[2].Amount)
		assert.Equal(t, 12.0, result.TotalSelected)
		assert.True(t, result.IsEnough)
	})

	// Test case 5: Not enough UTXOs
	t.Run("NotEnoughUtxos", func(t *testing.T) {
		result := SelectTokenUtxos(mockUtxos, 20.0, 2, nil)
		assert.Equal(t, mockUtxos, result.SelectedUtxos)
		assert.Equal(t, 15.0, result.TotalSelected)
		assert.False(t, result.IsEnough)
	})

	// Test case 6: Empty UTXOs
	t.Run("EmptyUtxos", func(t *testing.T) {
		result := SelectTokenUtxos([]*TokenUtxo{}, 5.0, 2, nil)
		assert.Empty(t, result.SelectedUtxos)
		assert.Equal(t, 0.0, result.TotalSelected)
		assert.False(t, result.IsEnough)
	})

	// Test case 7: Zero required amount
	t.Run("ZeroRequiredAmount", func(t *testing.T) {
		result := SelectTokenUtxos(mockUtxos, 0.0, 2, nil)
		assert.Equal(t, mockUtxos, result.SelectedUtxos)
		assert.Equal(t, 15.0, result.TotalSelected)
		assert.True(t, result.IsEnough)
	})

	// Test case 8: Different decimal places
	t.Run("DifferentDecimalPlaces", func(t *testing.T) {
		result := SelectTokenUtxos(mockUtxos, 0.000003, 6, nil)
		assert.Equal(t, 1, len(result.SelectedUtxos))
		assert.Equal(t, 0.0001, result.TotalSelected)
		assert.True(t, result.IsEnough)
	})

	// Test case 9: SmallestFirst input strategy and LargestFirst output strategy
	t.Run("SmallestFirstInputAndLargestFirstOutput", func(t *testing.T) {
		options := &TokenSelectionOptions{
			InputStrategy:  TokenSelectionStrategySmallestFirst,
			OutputStrategy: TokenSelectionStrategyLargestFirst,
		}
		result := SelectTokenUtxos(mockUtxos, 6.0, 2, options)
		assert.Equal(t, 3, len(result.SelectedUtxos))
		assert.Equal(t, mockUtxos[2].Amount, result.SelectedUtxos[0].Amount)
		assert.Equal(t, mockUtxos[1].Amount, result.SelectedUtxos[1].Amount)
		assert.Equal(t, mockUtxos[0].Amount, result.SelectedUtxos[2].Amount)
		assert.Equal(t, 6.0, result.TotalSelected)
		assert.True(t, result.IsEnough)
	})
}

// Test the ToToken and FromToken conversion functions
func TestTokenConversion(t *testing.T) {
	// Test ToToken
	t.Run("ToToken", func(t *testing.T) {
		assert.Equal(t, 1.0, ToToken(100, 2))
		assert.Equal(t, 0.5, ToToken(50, 2))
		assert.Equal(t, 1.0, ToToken(1000000, 6))
		assert.Equal(t, 0.0001, ToToken(100, 6))
	})

	// Test FromToken
	t.Run("FromToken", func(t *testing.T) {
		assert.Equal(t, uint64(100), FromToken(1.0, 2))
		assert.Equal(t, uint64(50), FromToken(0.5, 2))
		assert.Equal(t, uint64(1000000), FromToken(1.0, 6))
		assert.Equal(t, uint64(100), FromToken(0.0001, 6))
	})
}

func TestValidateSubTypeData(t *testing.T) {
	t.Run("NilData", func(t *testing.T) {
		result := ValidateSubTypeData(TokenTypeBSV21, "collection", nil)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "subtype data is nil")
	})

	t.Run("ValidCollection", func(t *testing.T) {
		data := &SubTypeData{
			Description: "Test Collection",
			Quantity:    100,
			Traits: map[string]interface{}{
				"trait1": "value1",
			},
		}
		result := ValidateSubTypeData(TokenTypeBSV21, "collection", data)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("InvalidCollection", func(t *testing.T) {
		data := &SubTypeData{
			// Missing description
			Quantity: 0, // Invalid quantity
		}
		result := ValidateSubTypeData(TokenTypeBSV21, "collection", data)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "description is required for collection subtype")
		assert.Contains(t, result.Errors, "quantity must be positive for collection subtype")
	})

	t.Run("ValidCollectionItem", func(t *testing.T) {
		data := &SubTypeData{
			CollectionID: "collection123",
			MintNumber:   1,
			Rank:         5,
		}
		result := ValidateSubTypeData(TokenTypeBSV21, "collectionItem", data)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("InvalidCollectionItem", func(t *testing.T) {
		data := &SubTypeData{
			// Missing collectionId
			MintNumber: -1, // Invalid mint number
			Rank:       -2, // Invalid rank
		}
		result := ValidateSubTypeData(TokenTypeBSV21, "collectionItem", data)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "collectionId is required for collectionItem subtype")
		assert.Contains(t, result.Errors, "mintNumber must be non-negative for collectionItem subtype")
		assert.Contains(t, result.Errors, "rank must be non-negative for collectionItem subtype")
	})

	t.Run("UnknownSubtype", func(t *testing.T) {
		data := &SubTypeData{}
		result := ValidateSubTypeData(TokenTypeBSV21, "unknownSubtype", data)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "unknown subtype: unknownSubtype")
	})

	t.Run("UnknownProtocol", func(t *testing.T) {
		data := &SubTypeData{}
		result := ValidateSubTypeData("UnknownProtocol", "collection", data)
		assert.False(t, result.Valid)
		assert.Contains(t, result.Errors, "unknown protocol: UnknownProtocol")
	})
}

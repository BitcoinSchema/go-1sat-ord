package ordinals

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// TokenSelectionStrategy represents different strategies for selecting token UTXOs
type TokenSelectionStrategy string

const (
	// TokenSelectionStrategySmallestFirst selects the smallest UTXOs first
	TokenSelectionStrategySmallestFirst TokenSelectionStrategy = "smallest"
	// TokenSelectionStrategyLargestFirst selects the largest UTXOs first
	TokenSelectionStrategyLargestFirst TokenSelectionStrategy = "largest"
	// TokenSelectionStrategyRetainOrder maintains the original order of UTXOs
	TokenSelectionStrategyRetainOrder TokenSelectionStrategy = "retain"
	// TokenSelectionStrategyRandom selects UTXOs randomly
	TokenSelectionStrategyRandom TokenSelectionStrategy = "random"
)

// TokenSelectionOptions represents options for token selection
type TokenSelectionOptions struct {
	// InputStrategy determines how token UTXOs are selected
	InputStrategy TokenSelectionStrategy
	// OutputStrategy determines how selected token UTXOs are ordered in the result
	OutputStrategy TokenSelectionStrategy
}

// TokenSelectionResult represents the result of a token selection operation
type TokenSelectionResult struct {
	// SelectedUtxos are the token UTXOs selected for the transaction
	SelectedUtxos []*TokenUtxo
	// TotalSelected is the total amount of tokens selected (in display format)
	TotalSelected float64
	// IsEnough indicates whether the selected amount meets the required amount
	IsEnough bool
}

// ToToken converts a token amount from raw format to display format
// It divides the raw amount by 10^decimals to get the display format
func ToToken(amount uint64, decimals uint8) float64 {
	// For simple number division, we just use the standard math package
	// Unlike the TypeScript implementation, Go doesn't support variable return types
	// and doesn't have a built-in bigint type for arbitrary precision
	divisor := math.Pow10(int(decimals))
	return float64(amount) / divisor
}

// FromToken converts a token amount from display format to raw format
// It multiplies the display amount by 10^decimals to get the raw format
func FromToken(amount float64, decimals uint8) uint64 {
	// Validate input
	if amount < 0 {
		// Similar to TypeScript, we should handle negative values, but we're simplifying
		// by not supporting them directly. Applications can handle sign logic.
		panic("FromToken cannot handle negative values directly")
	}

	// Check for potential overflow
	maxSafeValue := math.Pow(2, 64) / math.Pow10(int(decimals))
	if amount > maxSafeValue {
		panic(fmt.Sprintf("Value too large: %f exceeds maximum safe value of %f", amount, maxSafeValue))
	}

	multiplier := math.Pow10(int(decimals))
	// Round to nearest integer to handle floating point precision issues
	return uint64(math.Round(amount * multiplier))
}

// SelectTokenUtxos selects token UTXOs based on the required amount and specified strategies
// It returns the selected UTXOs, the total amount selected, and whether the selected amount is enough
func SelectTokenUtxos(
	tokenUtxos []*TokenUtxo,
	requiredTokens float64,
	decimals uint8,
	options *TokenSelectionOptions,
) *TokenSelectionResult {
	// Default options if none provided
	if options == nil {
		options = &TokenSelectionOptions{
			InputStrategy:  TokenSelectionStrategyRetainOrder,
			OutputStrategy: TokenSelectionStrategyRetainOrder,
		}
	}

	// Set default input strategy if not provided
	if options.InputStrategy == "" {
		options.InputStrategy = TokenSelectionStrategyRetainOrder
	}

	// Set default output strategy if not provided
	if options.OutputStrategy == "" {
		options.OutputStrategy = TokenSelectionStrategyRetainOrder
	}

	// Make a copy of the input UTXOs
	sortedUtxos := make([]*TokenUtxo, len(tokenUtxos))
	copy(sortedUtxos, tokenUtxos)

	// Sort the UTXOs based on the input strategy
	switch options.InputStrategy {
	case TokenSelectionStrategySmallestFirst:
		sort.Slice(sortedUtxos, func(i, j int) bool {
			return sortedUtxos[i].Amount < sortedUtxos[j].Amount
		})
	case TokenSelectionStrategyLargestFirst:
		sort.Slice(sortedUtxos, func(i, j int) bool {
			return sortedUtxos[i].Amount > sortedUtxos[j].Amount
		})
	case TokenSelectionStrategyRandom:
		// Shuffle the UTXOs
		rand.Shuffle(len(sortedUtxos), func(i, j int) {
			sortedUtxos[i], sortedUtxos[j] = sortedUtxos[j], sortedUtxos[i]
		})
	case TokenSelectionStrategyRetainOrder:
		// No sorting needed
	}

	// Select UTXOs until we have enough
	var totalSelected float64
	selectedUtxos := []*TokenUtxo{}

	for _, utxo := range sortedUtxos {
		selectedUtxos = append(selectedUtxos, utxo)
		totalSelected += ToToken(utxo.Amount, decimals)

		// Stop if we have enough (but only if requiredTokens > 0)
		if totalSelected >= requiredTokens && requiredTokens > 0 {
			break
		}
	}

	// Sort the selected UTXOs based on the output strategy
	if options.OutputStrategy != TokenSelectionStrategyRetainOrder {
		switch options.OutputStrategy {
		case TokenSelectionStrategySmallestFirst:
			sort.Slice(selectedUtxos, func(i, j int) bool {
				return selectedUtxos[i].Amount < selectedUtxos[j].Amount
			})
		case TokenSelectionStrategyLargestFirst:
			sort.Slice(selectedUtxos, func(i, j int) bool {
				return selectedUtxos[i].Amount > selectedUtxos[j].Amount
			})
		case TokenSelectionStrategyRandom:
			// Shuffle the UTXOs
			rand.Shuffle(len(selectedUtxos), func(i, j int) {
				selectedUtxos[i], selectedUtxos[j] = selectedUtxos[j], selectedUtxos[i]
			})
		}
	}

	return &TokenSelectionResult{
		SelectedUtxos: selectedUtxos,
		TotalSelected: totalSelected,
		IsEnough:      totalSelected >= requiredTokens,
	}
}

// SubTypeData represents metadata for various token subtypes
type SubTypeData struct {
	// Common fields that can appear in any subtype
	CollectionID string `json:"collectionId,omitempty"`
	Description  string `json:"description,omitempty"`
	Quantity     int    `json:"quantity,omitempty"`
	MintNumber   int    `json:"mintNumber,omitempty"`
	Rank         int    `json:"rank,omitempty"`
	RarityLabel  string `json:"rarityLabel,omitempty"`
	// Specific fields
	Traits map[string]interface{} `json:"traits,omitempty"`
}

// SubTypeValidationResult represents the result of validating subtype data
type SubTypeValidationResult struct {
	// Valid indicates whether the subtype data is valid
	Valid bool
	// Errors contains validation error messages if Valid is false
	Errors []string
}

// ValidateSubTypeData validates subtype data based on token protocol and subtype
func ValidateSubTypeData(protocol TokenType, subType string, data *SubTypeData) *SubTypeValidationResult {
	result := &SubTypeValidationResult{
		Valid:  true,
		Errors: []string{},
	}

	// Common validations for all subtypes
	if data == nil {
		result.Valid = false
		result.Errors = append(result.Errors, "subtype data is nil")
		return result
	}

	// Validate based on subtype
	switch subType {
	case "collection":
		// Required fields for collections
		if data.Description == "" {
			result.Valid = false
			result.Errors = append(result.Errors, "description is required for collection subtype")
		}
		if data.Quantity <= 0 {
			result.Valid = false
			result.Errors = append(result.Errors, "quantity must be positive for collection subtype")
		}
		// Validate traits if present
		if data.Traits != nil && len(data.Traits) > 0 {
			// In a real implementation, you would validate the structure of traits
			// For now, we just check that it's not empty
			if len(data.Traits) == 0 {
				result.Valid = false
				result.Errors = append(result.Errors, "traits must not be empty if present")
			}
		}

	case "collectionItem":
		// Required fields for collection items
		if data.CollectionID == "" {
			result.Valid = false
			result.Errors = append(result.Errors, "collectionId is required for collectionItem subtype")
		}
		// Validate mint number if present
		if data.MintNumber < 0 {
			result.Valid = false
			result.Errors = append(result.Errors, "mintNumber must be non-negative for collectionItem subtype")
		}
		// Validate rank if present
		if data.Rank < 0 {
			result.Valid = false
			result.Errors = append(result.Errors, "rank must be non-negative for collectionItem subtype")
		}

	default:
		// Unknown subtype
		result.Valid = false
		result.Errors = append(result.Errors, "unknown subtype: "+subType)
	}

	// Protocol-specific validations
	switch protocol {
	case TokenTypeBSV21:
		// BSV21 specific validations
		// (none for now, but could be added in the future)

	default:
		// Unknown protocol
		result.Valid = false
		result.Errors = append(result.Errors, "unknown protocol: "+string(protocol))
	}

	return result
}

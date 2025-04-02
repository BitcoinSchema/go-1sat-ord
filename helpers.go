package ordinals

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/bsv-blockchain/go-sdk/transaction"
)

const (
	// OneSatApiBase is the base URL for the 1Sat API
	OneSatApiBase = "https://ordinals.gorillapool.io/api/v1"
)

// UTXOResponse represents a UTXO response from the 1Sat API
type UTXOResponse struct {
	Txid   string `json:"txid"`
	Vout   int    `json:"vout"`
	Value  int    `json:"value"`
	Height int    `json:"height"`
	Script string `json:"script"`
}

// FetchPayUtxos fetches UTXOs for payment from the 1Sat API
func FetchPayUtxos(address string) ([]*Utxo, error) {
	url := fmt.Sprintf("%s/address/%s/utxo", OneSatApiBase, address)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch pay UTXOs: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var utxoResp []UTXOResponse
	if err := json.Unmarshal(body, &utxoResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	utxos := make([]*Utxo, 0, len(utxoResp))
	for _, u := range utxoResp {
		utxos = append(utxos, &Utxo{
			TxID:         u.Txid,
			Vout:         uint32(u.Vout),
			ScriptPubKey: u.Script,
			Satoshis:     uint64(u.Value),
		})
	}

	return utxos, nil
}

// NftUtxoResponse represents an NFT UTXO response from the 1Sat API
type NftUtxoResponse struct {
	UTXOResponse
	Origin      string `json:"origin"`
	ContentType string `json:"contentType"`
}

// FetchNftUtxos fetches NFT UTXOs from the 1Sat API
func FetchNftUtxos(address string, collectionID string) ([]*NftUtxo, error) {
	url := fmt.Sprintf("%s/address/%s/ordinals", OneSatApiBase, address)
	if collectionID != "" {
		url += "?collection=" + collectionID
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch NFT UTXOs: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var utxoResp []NftUtxoResponse
	if err := json.Unmarshal(body, &utxoResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	utxos := make([]*NftUtxo, 0, len(utxoResp))
	for _, u := range utxoResp {
		utxos = append(utxos, &NftUtxo{
			Utxo: Utxo{
				TxID:         u.Txid,
				Vout:         uint32(u.Vout),
				ScriptPubKey: u.Script,
				Satoshis:     uint64(u.Value),
			},
			ContentType:  u.ContentType,
			CollectionID: u.Origin,
		})
	}

	return utxos, nil
}

// TokenUtxoResponse represents a token UTXO response from the 1Sat API
type TokenUtxoResponse struct {
	UTXOResponse
	TokenID  string `json:"id"`
	Protocol string `json:"protocol"`
	Amount   string `json:"amount"`
	Decimals int    `json:"decimals"`
}

// FetchTokenUtxos fetches token UTXOs from the 1Sat API
func FetchTokenUtxos(protocol TokenType, tokenID string, address string) ([]*TokenUtxo, error) {
	url := fmt.Sprintf("%s/address/%s/tokens?protocol=%s", OneSatApiBase, address, protocol)
	if tokenID != "" {
		url += "&id=" + tokenID
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token UTXOs: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var utxoResp []TokenUtxoResponse
	if err := json.Unmarshal(body, &utxoResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	utxos := make([]*TokenUtxo, 0, len(utxoResp))
	for _, u := range utxoResp {
		amount, err := strconv.ParseUint(u.Amount, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse amount: %w", err)
		}

		utxos = append(utxos, &TokenUtxo{
			Utxo: Utxo{
				TxID:         u.Txid,
				Vout:         uint32(u.Vout),
				ScriptPubKey: u.Script,
				Satoshis:     uint64(u.Value),
			},
			TokenID:  u.TokenID,
			Protocol: TokenType(u.Protocol),
			Amount:   amount,
			Decimals: uint8(u.Decimals),
		})
	}

	return utxos, nil
}

// OneSatBroadcaster returns a function for broadcasting transactions using the 1Sat API
func OneSatBroadcaster() BroadcastFunc {
	return func(tx *transaction.Transaction) (*BroadcastResult, error) {
		url := fmt.Sprintf("%s/tx", OneSatApiBase)

		// Get the transaction hex
		txHex := tx.String()

		// Create the request body
		reqBody := map[string]string{"rawtx": txHex}
		reqJSON, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		// Make the request
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqJSON))
		if err != nil {
			return nil, fmt.Errorf("failed to broadcast transaction: %w", err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				fmt.Printf("Error closing response body: %v\n", err)
			}
		}()

		// Parse the response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		// Check for success
		if resp.StatusCode != http.StatusOK {
			return &BroadcastResult{
				Status:  "error",
				Message: string(body),
			}, nil
		}

		// Return the txid as success
		return &BroadcastResult{
			Status: "success",
			TxID:   tx.TxID().String(),
		}, nil
	}
}

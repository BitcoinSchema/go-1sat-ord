package ordinals

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock RoundTripper that redirects requests to our test server
type mockRoundTripper struct {
	serverURL string
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the URL to point to our test server
	req.URL.Scheme = "http"
	req.URL.Host = m.serverURL[7:] // remove http://

	// Use the standard transport to do the actual request
	return http.DefaultTransport.RoundTrip(req)
}

func TestFetchPayUtxos(t *testing.T) {
	// Create a test server with handler
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the correct endpoint is being called
		assert.Equal(t, "/api/v1/address/test_address/utxo", r.URL.Path)

		// Return a sample response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`[
			{
				"txid": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				"vout": 0,
				"value": 100000,
				"height": 123456,
				"script": "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac"
			}
		]`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	// Override standard HTTP client with our test client for this test
	origClient := http.DefaultClient
	http.DefaultClient = &http.Client{
		Transport: &mockRoundTripper{server.URL},
	}
	defer func() { http.DefaultClient = origClient }()

	// Call the function
	utxos, err := FetchPayUtxos("test_address")

	// Check the result
	assert.NoError(t, err)
	assert.Len(t, utxos, 1)
	assert.Equal(t, "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", utxos[0].TxID)
	assert.Equal(t, uint32(0), utxos[0].Vout)
	assert.Equal(t, uint64(100000), utxos[0].Satoshis)
	assert.Equal(t, "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac", utxos[0].ScriptPubKey)
}

func TestFetchNftUtxos(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the correct endpoint is being called
		assert.Equal(t, "/api/v1/address/test_address/ordinals", r.URL.Path)
		assert.Equal(t, "collection=test_collection", r.URL.RawQuery)

		// Return a sample response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`[
			{
				"txid": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				"vout": 0,
				"value": 1,
				"height": 123456,
				"script": "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
				"origin": "test_collection",
				"contentType": "text/plain"
			}
		]`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	// Override standard HTTP client with our test client for this test
	origClient := http.DefaultClient
	http.DefaultClient = &http.Client{
		Transport: &mockRoundTripper{server.URL},
	}
	defer func() { http.DefaultClient = origClient }()

	// Call the function
	utxos, err := FetchNftUtxos("test_address", "test_collection")

	// Check the result
	assert.NoError(t, err)
	assert.Len(t, utxos, 1)
	assert.Equal(t, "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", utxos[0].TxID)
	assert.Equal(t, uint32(0), utxos[0].Vout)
	assert.Equal(t, uint64(1), utxos[0].Satoshis)
	assert.Equal(t, "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac", utxos[0].ScriptPubKey)
	assert.Equal(t, "text/plain", utxos[0].ContentType)
	assert.Equal(t, "test_collection", utxos[0].CollectionID)
}

func TestFetchTokenUtxos(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the correct endpoint is being called
		assert.Equal(t, "/api/v1/address/test_address/tokens", r.URL.Path)
		assert.Equal(t, "protocol=bsv-21&id=test_token", r.URL.RawQuery)

		// Return a sample response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`[
			{
				"txid": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
				"vout": 0,
				"value": 1,
				"height": 123456,
				"script": "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac",
				"id": "test_token",
				"protocol": "bsv-21",
				"amount": "1000",
				"decimals": 0
			}
		]`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	// Override standard HTTP client with our test client for this test
	origClient := http.DefaultClient
	http.DefaultClient = &http.Client{
		Transport: &mockRoundTripper{server.URL},
	}
	defer func() { http.DefaultClient = origClient }()

	// Call the function
	utxos, err := FetchTokenUtxos(TokenTypeBSV21, "test_token", "test_address")

	// Check the result
	assert.NoError(t, err)
	assert.Len(t, utxos, 1)
	assert.Equal(t, "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890", utxos[0].TxID)
	assert.Equal(t, uint32(0), utxos[0].Vout)
	assert.Equal(t, uint64(1), utxos[0].Satoshis)
	assert.Equal(t, "76a914b437a081c28a178b9ce5e2a0e694d45d1d5e2c0388ac", utxos[0].ScriptPubKey)
	assert.Equal(t, "test_token", utxos[0].TokenID)
	assert.Equal(t, TokenTypeBSV21, utxos[0].Protocol)
	assert.Equal(t, uint64(1000), utxos[0].Amount)
	assert.Equal(t, uint8(0), utxos[0].Decimals)
}

func TestOneSatBroadcaster(t *testing.T) {
	// This is a mock test since we can't easily test with a real transaction
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the correct endpoint is being called
		assert.Equal(t, "/api/v1/tx", r.URL.Path)

		// Return a sample response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(`{"txid": "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"}`))
		assert.NoError(t, err)
	}))
	defer server.Close()

	// We can't easily create a real transaction for testing
	// Just verify that the function returns a broadcast function
	broadcaster := OneSatBroadcaster()
	assert.NotNil(t, broadcaster)
}

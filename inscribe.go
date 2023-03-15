package ordinals

import (
	"encoding/hex"

	"github.com/bitcoinschema/go-aip"
	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bt/v2"
)

// Ordinal is the ordinal inscription data
type Ordinal struct {
	Data        []byte
	ContentType string
}

// ORD is the inscription protocol prefix
const Prefix = "ord"

// Inscribe creates a 1sat ordinal output immediately followed by an inscription output.
// utxos - unspend outputs for payment
// opReturn - data to be inscribed
// pursePk - private key for signing inputs
// changeAddress - where torecieve change
// tokenAddress - where to recieve the ordinal
// signingAddress - key to use when signing the inscription data
// signingKey - private key to use for signing inscription data
func Inscribe(utxos []*bitcoin.Utxo, inscriptionData *Ordinal, opReturn bitcoin.OpReturnData, pursePk *bec.PrivateKey, changeAddress string, tokenAddress string, signingAddress *string, signingKey *string) (tx *bt.Tx, err error) {

	payToAddresses := []*bitcoin.PayToAddress{{Address: tokenAddress, Satoshis: 1}}

	// Sign with AIP
	_, outData, _, err := aip.SignOpReturnData(*signingKey, "BITCOIN_ECDSA", opReturn)
	if err != nil {
		return nil, err
	}

	// Create ASM from data
	var opReturnAsm string
	for _, push := range outData {
		pushHex := hex.EncodeToString(push)
		opReturnAsm = opReturnAsm + " " + pushHex
	}
	// Create Inscription Tx
	tx, err = CreateTxWithChange(utxos, payToAddresses, inscriptionData, &opReturnAsm, changeAddress, nil, nil, pursePk, true)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

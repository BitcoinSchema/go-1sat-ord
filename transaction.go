package ordinals

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bt/v2"
	"github.com/libsv/go-bt/v2/bscript"
	"github.com/libsv/go-bt/v2/unlocker"
)

// account is a struct/interface for implementing unlocker
type account struct {
	PrivateKey *bec.PrivateKey
}

// Unlocker get the correct un-locker for a given locking script
func (a *account) Unlocker(context.Context, *bscript.Script) (bt.Unlocker, error) {
	return &unlocker.Simple{
		PrivateKey: a.PrivateKey,
	}, nil
}

func CreateTx(utxos []*bitcoin.Utxo, addresses []*bitcoin.PayToAddress, inscriptionData *Ordinal,
	opReturnAsm *string, privateKey *bec.PrivateKey) (*bt.Tx, error) {

	// Start creating a new transaction
	tx := bt.NewTx()

	// Accumulate the total satoshis from all utxo(s)
	var totalSatoshis uint64

	// Loop all utxos and add to the transaction
	var err error
	for _, utxo := range utxos {
		if err = tx.From(utxo.TxID, utxo.Vout, utxo.ScriptPubKey, utxo.Satoshis); err != nil {
			return nil, err
		}
		totalSatoshis += utxo.Satoshis
	}

	// Loop any pay to addresses
	for _, address := range addresses {
		var a *bscript.Script
		a, err = bscript.NewP2PKHFromAddress(address.Address)
		if err != nil {
			return nil, err
		}

		// Handle Ordinals
		if address.Satoshis == 1 && inscriptionData != nil {
			// 1sat ordinals prefix "1sat"

			inscriptionHex := hex.EncodeToString(inscriptionData.Data)
			inscriptionContentTypeHex := hex.EncodeToString([]byte(inscriptionData.ContentType))
			ordHex := hex.EncodeToString([]byte("ord"))

			ordAsm := "OP_FALSE OP_IF " + ordHex + " OP_1 " + inscriptionContentTypeHex + " OP_0 " + inscriptionHex + " OP_ENDIF"
			if opReturnAsm != nil {
				ordAsm = ordAsm + " OP_RETURN " + *opReturnAsm
			}
			aAsm, err := a.ToASM()
			if err != nil {
				return nil, err
			}
			// 1sat OP_DROP P2PKH
			newAsm := aAsm + " " + ordAsm
			newA, err := bscript.NewFromASM(newAsm)
			if err != nil {
				return nil, err
			}
			a = newA
		}
		tx.AddOutput(&bt.Output{
			Satoshis:      address.Satoshis,
			LockingScript: a,
		})
	}

	// If inputs are supplied, make sure they are sufficient for this transaction
	if len(tx.Inputs) > 0 {
		// Sanity check - not enough satoshis in utxo(s) to cover all paid amount(s)
		// They should never be equal, since the fee is the spread between the two amounts
		totalOutputSatoshis := tx.TotalOutputSatoshis() // Does not work properly
		if totalOutputSatoshis > totalSatoshis {
			return nil, fmt.Errorf("not enough in utxo(s) to cover: %d + (fee) found: %d", totalOutputSatoshis, totalSatoshis)
		}
	}

	// Sign the transaction
	if privateKey != nil {
		myAccount := &account{PrivateKey: privateKey}
		// todo: support context (ctx)
		if err = tx.FillAllInputs(context.Background(), myAccount); err != nil {
			return nil, err
		}
	}

	// Return the transaction as a raw string
	return tx, nil
}

func CreateTxWithChange(utxos []*bitcoin.Utxo, payToAddresses []*bitcoin.PayToAddress, inscriptionData *Ordinal, opReturnAsm *string,
	changeAddress string, standardRate, dataRate *bt.Fee,
	privateKey *bec.PrivateKey, sendingOrdinal bool) (*bt.Tx, error) {

	// Missing utxo(s) or change address
	if len(utxos) == 0 {
		return nil, fmt.Errorf("UTXOs required")
	} else if len(changeAddress) == 0 {
		return nil, fmt.Errorf("change address required")
	}

	// Accumulate the total satoshis from all utxo(s)
	var totalSatoshis uint64
	var totalPayToSatoshis uint64
	var remainder uint64
	var hasChange bool

	// Loop utxos and get total usable satoshis
	for _, utxo := range utxos {
		totalSatoshis += utxo.Satoshis
	}

	// Loop all payout address amounts
	for _, address := range payToAddresses {
		totalPayToSatoshis += address.Satoshis
	}

	// Sanity check - already not enough satoshis?
	if totalPayToSatoshis > totalSatoshis {
		return nil, fmt.Errorf(
			"not enough in utxo(s) to cover: %d + (fee), total found: %d",
			totalPayToSatoshis,
			totalSatoshis,
		)
	}

	// Add the change address as the difference (all change except 1 sat for Draft tx)
	// Only if the tx is NOT for the full amount
	if totalPayToSatoshis != totalSatoshis {
		hasChange = true
		payToAddresses = append(payToAddresses, &bitcoin.PayToAddress{
			Address:  changeAddress,
			Satoshis: totalSatoshis - (totalPayToSatoshis + 1),
		})
	}

	// Create the "Draft tx"
	fee, err := draftTx(utxos, payToAddresses, inscriptionData, opReturnAsm, privateKey, standardRate, dataRate)
	if err != nil {
		return nil, err
	}

	// Check that we have enough to cover the fee
	if (totalPayToSatoshis + fee) > totalSatoshis {

		// Remove temporary change address first
		if hasChange {
			payToAddresses = payToAddresses[:len(payToAddresses)-1]
		}

		// Re-run draft tx with no change address
		if fee, err = draftTx(
			utxos, payToAddresses, inscriptionData, opReturnAsm, privateKey, standardRate, dataRate,
		); err != nil {
			return nil, err
		}

		// Get the remainder missing (handle negative overflow safer)
		totalToPay := totalPayToSatoshis + fee
		if totalToPay >= totalSatoshis {
			remainder = totalToPay - totalSatoshis
		} else {
			remainder = totalSatoshis - totalToPay
		}

		// Remove remainder from last used payToAddress (or continue until found)
		feeAdjusted := false
		for i := len(payToAddresses) - 1; i >= 0; i-- { // Working backwards
			if payToAddresses[i].Satoshis > remainder {
				payToAddresses[i].Satoshis = payToAddresses[i].Satoshis - remainder
				feeAdjusted = true
				break
			}
		}

		// Fee was not adjusted (all inputs do not cover the fee)
		if !feeAdjusted {
			return nil, fmt.Errorf(
				"auto-fee could not be applied without removing an output (payTo %d) "+
					"(amount %d) (remainder %d) (fee %d) (total %d)",
				len(payToAddresses), totalPayToSatoshis, remainder, fee, totalSatoshis,
			)
		}

	} else {

		// Remove the change address (old version with original satoshis)
		// Add the change address as the difference (now with adjusted fee)
		if hasChange {
			payToAddresses = payToAddresses[:len(payToAddresses)-1]

			payToAddresses = append(payToAddresses, &bitcoin.PayToAddress{
				Address:  changeAddress,
				Satoshis: totalSatoshis - (totalPayToSatoshis + fee),
			})
		}
	}

	// Create the "Final tx" (or error)
	return CreateTx(utxos, payToAddresses, inscriptionData, opReturnAsm, privateKey)
}

// draftTx is a helper method to create a draft tx and associated fees
func draftTx(utxos []*bitcoin.Utxo, payToAddresses []*bitcoin.PayToAddress, inscriptionData *Ordinal, opReturnAsm *string,
	privateKey *bec.PrivateKey, standardRate, dataRate *bt.Fee) (uint64, error) {

	// Create the "Draft tx"
	tx, err := CreateTx(utxos, payToAddresses, inscriptionData, opReturnAsm, privateKey)
	if err != nil {
		return 0, err
	}

	// Calculate the fees for the "Draft tx"
	// todo: hack to add 1 extra sat - ensuring that fee is over the minimum with rounding issues in WOC and other systems
	fee := bitcoin.CalculateFeeForTx(tx, standardRate, dataRate) + 1
	return fee, nil
}

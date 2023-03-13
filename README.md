# go-1sat-ord

A package for creating 1sat ordinals on Bitcoin SV.

### Usage

```go
import (
  github.com/bitcoinschema/go-1sat-ord
)
// utxos available to use for transaction fee, and source of the 1sat ordinal
var utxos []*bitcoin.Utxo
// inscriptionData used to create an inscription output script
var inscriptionData ordinals.Ordinal
// (optional) opReturn array to be added after inscription
var opReturn bitcoin.OpReturnData
// a private key in wif format for signing the utxos and funding the tx
var purseWif string
// address to return remaining funds after fees and 1sat
var changeAddress string
// destination address - will recieve the 1sat ordinal
var ordinalAddress string

tx, err := ordinals.Inscribe(utxos, inscriptionData, opReturn, purseWif, changeAddress, ordinalAddress, signingAddress, signingKey)

// tx is a *bt.Tx from bsvlib/go-bt
// tx.TxID()
```

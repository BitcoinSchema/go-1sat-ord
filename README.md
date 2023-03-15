# go-1sat-ord

A package for creating 1sat Ordinals on Bitcoin SV.

### Configuration

Rename `sample.env` to `.env` and set the wif keys.

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

## More Information

[1Sat Ordinals](https://github.com/bitcoinschema/1sat-ordinals)

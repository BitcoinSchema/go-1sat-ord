# go-1sat-ord

A package for creating 1sat ordinals on Bitcoin SV.

### Dependencies

```go
github.com/bitcoinschema/go-bitcoin
github.com/bitcoinschema/go-b
github.com/libsv/go-bt/v2
```

### Usage

```go
import (
  github.com/bitcoinschema/go-1sat-ord
)
// utxos available to use for transaction fee, and source of the 1sat ordinal
var utxos []*bitcoin.Utxo
// opReturn array to be inscribed. will be prefixed with ord inscription identifier "ord"
var opReturn bitcoin.OpReturnData
// a privat key in wif format for signing the utxos and funding the tx
var purseWif string
// address to return remaining funds after fees and 1sat
var changeAddress string
// destination address - will recieve the 1sat ordinal
var ordinalAddress string

ordinals.Inscribe(utxos, opReturn, purseWif, changeAddress, ordinalAddress, signingAddress, signingKey) (inscription *Inscription, tx *bt.Tx, err error) {

}
```

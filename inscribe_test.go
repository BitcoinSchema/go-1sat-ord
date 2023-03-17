package ordinals

import (
	"encoding/base64"
	"log"
	"testing"

	"github.com/bitcoinschema/go-bitcoin/v2"
	"github.com/bitcoinschema/go-bmap"
	magic "github.com/bitcoinschema/go-map"
)

func TestInscription(t *testing.T) {
	// gltf-binary base64 encoded
	fireShard := "Z2xURgIAAABobgwAbAcAAEpTT057ImFzc2V0Ijp7ImdlbmVyYXRvciI6Ik1pY3Jvc29mdCBHTFRGIEV4cG9ydGVyIDIuOC4zLjQwIiwidmVyc2lvbiI6IjIuMCJ9LCJhY2Nlc3NvcnMiOlt7ImJ1ZmZlclZpZXciOjAsImNvbXBvbmVudFR5cGUiOjUxMjUsImNvdW50IjozNDgsInR5cGUiOiJTQ0FMQVIifSx7ImJ1ZmZlclZpZXciOjEsImNvbXBvbmVudFR5cGUiOjUxMjYsImNvdW50IjozNDgsInR5cGUiOiJWRUMzIiwibWF4IjpbMC4zNjIyNjQ5OTA4MDY1Nzk2LDEuMTU1ODU4OTkzNTMwMjczNSwwLjQwOTIwNjAwMjk1MDY2ODM2XSwibWluIjpbLTAuNDUyMjI4OTkzMTc3NDEzOTYsLTAuNjI1MDc5OTg5NDMzMjg4NiwtMC40MDY1NDk5OTAxNzcxNTQ1Nl19LHsiYnVmZmVyVmlldyI6MiwiY29tcG9uZW50VHlwZSI6NTEyNiwiY291bnQiOjM0OCwidHlwZSI6IlZFQzMifSx7ImJ1ZmZlclZpZXciOjMsImNvbXBvbmVudFR5cGUiOjUxMjYsImNvdW50IjozNDgsInR5cGUiOiJWRUM0In0seyJidWZmZXJWaWV3Ijo0LCJjb21wb25lbnRUeXBlIjo1MTI2LCJjb3VudCI6MzQ4LCJ0eXBlIjoiVkVDMiJ9XSwiYnVmZmVyVmlld3MiOlt7ImJ1ZmZlciI6MCwiYnl0ZU9mZnNldCI6MCwiYnl0ZUxlbmd0aCI6MTM5MiwidGFyZ2V0IjozNDk2M30seyJidWZmZXIiOjAsImJ5dGVPZmZzZXQiOjEzOTIsImJ5dGVMZW5ndGgiOjQxNzYsInRhcmdldCI6MzQ5NjJ9LHsiYnVmZmVyIjowLCJieXRlT2Zmc2V0Ijo1NTY4LCJieXRlTGVuZ3RoIjo0MTc2LCJ0YXJnZXQiOjM0OTYyfSx7ImJ1ZmZlciI6MCwiYnl0ZU9mZnNldCI6OTc0NCwiYnl0ZUxlbmd0aCI6NTU2OCwidGFyZ2V0IjozNDk2Mn0seyJidWZmZXIiOjAsImJ5dGVPZmZzZXQiOjE1MzEyLCJieXRlTGVuZ3RoIjoyNzg0LCJ0YXJnZXQiOjM0OTYyfSx7ImJ1ZmZlciI6MCwiYnl0ZU9mZnNldCI6MTgwOTYsImJ5dGVMZW5ndGgiOjc5NDY2OX1dLCJidWZmZXJzIjpbeyJieXRlTGVuZ3RoIjo4MTI3NjV9XSwiaW1hZ2VzIjpbeyJidWZmZXJWaWV3Ijo1LCJtaW1lVHlwZSI6ImltYWdlL3BuZyJ9XSwibWF0ZXJpYWxzIjpbeyJwYnJNZXRhbGxpY1JvdWdobmVzcyI6eyJiYXNlQ29sb3JUZXh0dXJlIjp7ImluZGV4IjowfSwibWV0YWxsaWNGYWN0b3IiOjAuMCwicm91Z2huZXNzRmFjdG9yIjowLjEwOTgwMzkyOTkyNDk2NDl9LCJkb3VibGVTaWRlZCI6dHJ1ZX1dLCJtZXNoZXMiOlt7InByaW1pdGl2ZXMiOlt7ImF0dHJpYnV0ZXMiOnsiVEFOR0VOVCI6MywiTk9STUFMIjoyLCJQT1NJVElPTiI6MSwiVEVYQ09PUkRfMCI6NH0sImluZGljZXMiOjAsIm1hdGVyaWFsIjowfV19XSwibm9kZXMiOlt7ImNoaWxkcmVuIjpbMV0sInJvdGF0aW9uIjpbLTAuNzA3MTA2NzA5NDgwMjg1NiwwLjAsLTAuMCwwLjcwNzEwNjgyODY4OTU3NTJdLCJzY2FsZSI6WzEuMCwwLjk5OTk5OTk0MDM5NTM1NTIsMC45OTk5OTk5NDAzOTUzNTUyXSwibmFtZSI6IlJvb3ROb2RlIChnbHRmIG9yaWVudGF0aW9uIG1hdHJpeCkifSx7ImNoaWxkcmVuIjpbMl0sIm5hbWUiOiJSb290Tm9kZSAobW9kZWwgY29ycmVjdGlvbiBtYXRyaXgpIn0seyJjaGlsZHJlbiI6WzNdLCJyb3RhdGlvbiI6WzAuNzA3MTA2NzY5MDg0OTMwNCwwLjAsMC4wLDAuNzA3MTA2NzY5MDg0OTMwNF0sIm5hbWUiOiI5MzcwMjFkNDkyYjM0ZjBlOWM4NDU3YjBmMTNhYTBmZS5mYngifSx7ImNoaWxkcmVuIjpbNF0sIm5hbWUiOiJSb290Tm9kZSJ9LHsiY2hpbGRyZW4iOls1XSwibmFtZSI6ImNyeXN0YWxMb3c6TWVzaCJ9LHsibWVzaCI6MCwibmFtZSI6ImNyeXN0YWxMb3c6TWVzaF9sYW1iZXJ0NF8wIn1dLCJzYW1wbGVycyI6W3sibWFnRmlsdGVyIjo5NzI5LCJtaW5GaWx0ZXIiOjk5ODd9XSwic2NlbmVzIjpbeyJub2RlcyI6WzBdfV0sInRleHR1cmVzIjpbeyJzYW1wbGVyIjowLCJzb3VyY2UiOjB9XSwic2NlbmUiOjB94GYMAEJJTgAAAAAAAQAAAAIAAAADAAAABAAAAAUAAAAGAAAABwAAAAgAAAAJAAAACgAAAAsAAAAMAAAADQAAAA4AAAAPAAAAEAAAABEAAAASAAAAEwAAABQAAAAVAAAAFgAAABcAAAAYAAAAGQAAABoAAAAbAAAAHAAAAB0AAAAeAAAAHwAAACAAAAAhAAAAIgAAACMAAAAkAAAAJQAAACYAAAAnAAAAKAAAACkAAAAqAAAAKwAAACwAAAAtAAAALgAAAC8AAAAwAAAAMQAAADIAAAAzAAAANAAAADUAAAA2AAAANwAAADgAAAA5AAAAOgAAADsAAAA8AAAAPQAAAD4AAAA"
	var fireShardMediaType = "model/gltf-binary"

	paymentWif := "L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw"
	// The data signature to use when inscribing
	signingWif := "L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw"
	// The key of the wallet to hold the 1sat ordinals
	ordinalWif := "L3MhnEn1pLWcggeYLk9jdkvA2wUK1iWwwrGkBbgQRqv6HPCdRxuw"
	// Get satchel address
	ordinalPk, err := bitcoin.WifToPrivateKey(ordinalWif)
	if err != nil {
		t.Fail()
		return
	}

	ordinalAddress, err := bitcoin.GetAddressFromPrivateKey(ordinalPk, true)
	if err != nil {
		t.Fail()
		return
	}

	// Get change address
	changePk, err := bitcoin.WifToPrivateKey(paymentWif)
	if err != nil {
		t.Fail()
		return
	}

	changeAddress, err := bitcoin.GetAddressFromPrivateKey(changePk, true)
	if err != nil {
		t.Fail()
		return
	}

	// Get signing address
	signPk, err := bitcoin.WifToPrivateKey(signingWif)
	if err != nil {
		t.Fail()
		return
	}

	signingAddress, err := bitcoin.GetAddressFromPrivateKey(signPk, true)
	if err != nil {
		t.Fail()
		return
	}
	log.Println("change address", changeAddress)
	// Get a balance for an address
	fsBytes, err := base64.RawStdEncoding.DecodeString(fireShard)
	if err != nil {
		log.Fatal("error:", err)
	}

	// Set ordinal payload
	inscriptionData := &Ordinal{
		Data:        fsBytes,
		ContentType: fireShardMediaType,
	}

	inscribeUtxos := []*bitcoin.Utxo{{
		Satoshis:     271902,
		ScriptPubKey: "76a914199652f095a99e31d487743b9bd21ecc03343f4b88ac",
		TxID:         "28302ee1ac9c62b0098ee741e4cffad2798a6bbbd2f53467d6e7ca9829b1ac2a",
		Vout:         0,
	}}

	signingKey, err := bitcoin.WifToPrivateKeyString(signingWif)
	if err != nil {
		t.Fatal("Failed to create signing key")
	}
	paymentPk, err := bitcoin.WifToPrivateKey(paymentWif)
	if err != nil {
		t.Fatal("Failed to create payment key")
	}
	opReturn := bitcoin.OpReturnData{
		[]byte(magic.Prefix),
		[]byte(magic.Set),
		[]byte("type"),
		[]byte("post"),
		[]byte("context"),
		[]byte("geohash"),
		[]byte("geohash"),
		[]byte("dhxnd1pwn"),
	}
	tx, err := Inscribe(inscribeUtxos, inscriptionData, opReturn, paymentPk, changeAddress, ordinalAddress, &signingAddress, &signingKey)
	if err != nil {
		t.Fatalf("Inscription failed %s", err)
	}

	rawTx := tx.String()

	bmapTx, err := bmap.NewFromTx(rawTx)
	if err != nil {
		t.Fatalf("error %s", err)
	}
	if len(bmapTx.Ord) != 1 {
		t.Fatalf("tx %+v, bmapTx: %+v", tx, bmapTx)
	}

}

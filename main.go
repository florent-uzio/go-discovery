package main

import (
	"encoding/hex"
	"fmt"

	binarycodec "github.com/xyield/xrpl-go/binary-codec"
	"github.com/xyield/xrpl-go/client/websocket"
	"github.com/xyield/xrpl-go/keypairs"
)

const ServerUrl = "wss://s.altnet.rippletest.net:51233/"
const AccountSeed = "sEdSmpmt3DbU42EgsxVrToBqcA6bn3P" // r3Z8KLzomveD5mu7tYsuhXDPJTfH91ZzHJ
const DestinationAddress = "rpNvHzENX8S2Bgx4dyCEcpqrb912QvRdoz"

type SubmitRequest struct {
	TxBlob   string `json:"tx_blob"`
	FailHard bool   `json:"fail_hard,omitempty"`
}

func (SubmitRequest) Validate() error {
	return nil
}
func (SubmitRequest) Method() string {
	return "submit"
}

/**
 * Example to send a payment
 */
func main() {
	privKey, pubKey, _ := keypairs.DeriveKeypair(AccountSeed, false)
	address, _ := keypairs.DeriveClassicAddress(pubKey)

	fmt.Printf("Sender Account %v\n", address)

	tx := map[string]any{
		"Account":         address,
		"TransactionType": "Payment",
		"Amount":          "20",
		"Destination":     DestinationAddress,
		"Flags":           0,
		"Fee":             "12",
		"Sequence":        1801574,
		"SigningPubKey":   pubKey,
	}

	encodedTx, _ := binarycodec.EncodeForSigning(tx)
	hexTx, err := hex.DecodeString(encodedTx)
	signature, _ := keypairs.Sign(string(hexTx), privKey)

	tx["TxnSignature"] = signature
	signedTx, err := binarycodec.Encode(tx)
	if err != nil {
		return
	}

	ws := websocket.NewWebsocketClient(&websocket.WebsocketConfig{
		URL: ServerUrl,
	})
	res, _ := ws.SendRequest(
		SubmitRequest{
			signedTx,
			false,
		})

	fmt.Printf("Tx Result %+v\n", res)
}

package main

import (
	"fmt"

	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl"
	"github.com/Peersyst/xrpl-go/xrpl/client/websocket"
	"github.com/Peersyst/xrpl-go/xrpl/faucet"
	"github.com/Peersyst/xrpl-go/xrpl/model/transactions"
	"github.com/Peersyst/xrpl-go/xrpl/model/transactions/types"
	// "github.com/Peersyst/xrpl-go/xrpl/utils" // use utils.CurrencyStringToHex(...) if your token has more than 3 characters
)

const (
	currencyCode = "FOO"
)

func main() {
	//
	// Configure client
	//
	fmt.Println("⏳ Setting up client...")
	client := websocket.NewWebsocketClient(
		websocket.NewWebsocketClientConfig().
			WithHost("wss://s.altnet.rippletest.net").
			WithFaucetProvider(faucet.NewTestnetFaucetProvider()),
	)
	fmt.Println("✅ Client configured!")
	fmt.Println()

	//
	// Configure wallets
	//
	fmt.Println("⏳ Setting up wallets...")
	issuerWallet, err := xrpl.NewWallet(addresscodec.ED25519)
	if err != nil {
		fmt.Printf("❌ Error creating issuer wallet: %s\n", err)
		return
	}
	err = client.FundWallet(&issuerWallet)
	if err != nil {
		fmt.Printf("❌ Error funding issuer wallet: %s\n", err)
		return
	}
	fmt.Println("💸 Issuer wallet funded!")

	customerOneWallet, err := xrpl.NewWallet(addresscodec.ED25519)
	if err != nil {
		fmt.Printf("❌ Error creating customer one wallet: %s\n", err)
		return
	}
	err = client.FundWallet(&customerOneWallet)
	if err != nil {
		fmt.Printf("❌ Error funding customer one wallet: %s\n", err)
		return
	}
	fmt.Println("💸 Customer one wallet funded!")

	customerTwoWallet, err := xrpl.NewWallet(addresscodec.ED25519)
	if err != nil {
		fmt.Printf("❌ Error creating token wallet: %s\n", err)
		return
	}
	err = client.FundWallet(&customerTwoWallet)
	if err != nil {
		fmt.Printf("❌ Error funding customer two wallet: %s\n", err)
		return
	}
	fmt.Println("💸 Customer two wallet funded!")
	fmt.Println()

	fmt.Println("✅ Wallets setup complete!")
	fmt.Println("💳 Issuer wallet:", issuerWallet.ClassicAddress)
	fmt.Println("💳 Customer one wallet:", customerOneWallet.ClassicAddress)
	fmt.Println("💳 Customer two wallet:", customerTwoWallet.ClassicAddress)
	fmt.Println()

	// **********************************
	// Configure issuing address settings
	// **********************************

	fmt.Println("⏳ Configuring issuing address settings...")
	issuingWalletAccountSet := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(issuerWallet.ClassicAddress),
		},
		TickSize:     5,
		TransferRate: 0,
		Domain:       "6578616D706C652E636F6D", // example.com
	}

	issuingWalletAccountSet.SetAsfDefaultRipple()
	issuingWalletAccountSet.SetDisallowXRP()
	// coldWalletAccountSet.SetAsfDepositAuth() // Potentially needed according to your needs

	flattenedTx := issuingWalletAccountSet.Flatten()

	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err := issuerWallet.Sign(flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error signing transaction: %s\n", err)
		return
	}

	response, err := client.SubmitTransactionBlob(txBlob, false)
	if err != nil {
		fmt.Printf("❌ Error submitting transaction: %s\n", err)
		return
	}

	if response.EngineResult != "tesSUCCESS" {
		fmt.Println("❌ Cold address settings configuration failed!", response.EngineResult)
		fmt.Println("Try again!")
		fmt.Println()
		return
	}

	fmt.Println("✅ Issuing address settings configured!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	// ********************************************************************
	// Create trust line from customer one to issuer address
	// ********************************************************************

	fmt.Println("⏳ Creating trust line from customer one to issuer address...")
	customerOneTrustSet := &transactions.TrustSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(customerOneWallet.ClassicAddress),
		},
		LimitAmount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(issuerWallet.ClassicAddress),
			Value:    "100000000000000",
		},
	}

	flattenedTx = customerOneTrustSet.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = customerOneWallet.Sign(flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error signing transaction: %s\n", err)
		return
	}

	response, err = client.SubmitTransactionBlob(txBlob, false)
	if err != nil {
		fmt.Printf("❌ Error submitting transaction: %s\n", err)
		return
	}

	if response.EngineResult != "tesSUCCESS" {
		fmt.Println("❌ Trust line from customer one to issuer address creation failed!", response.EngineResult)
		fmt.Println("Try again!")
		fmt.Println()
		return
	}

	fmt.Println("✅ Trust line from customer one to issuer address created!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	//
	// Create trust line from costumer two to issuer address
	//
	fmt.Println("⏳ Creating trust line from customer two to issuer address...")
	customerTwoTrustSet := &transactions.TrustSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(customerTwoWallet.ClassicAddress),
		},
		LimitAmount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(issuerWallet.ClassicAddress),
			Value:    "100000000000000",
		},
	}

	flattenedTx = customerTwoTrustSet.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = customerTwoWallet.Sign(flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error signing transaction: %s\n", err)
		return
	}

	response, err = client.SubmitTransactionBlob(txBlob, false)
	if err != nil {
		fmt.Printf("❌ Error submitting transaction: %s\n", err)
		return
	}

	if response.EngineResult != "tesSUCCESS" {
		fmt.Println("❌ Trust line from customer one to cold address creation failed!", response.EngineResult)
		fmt.Println("Try again!")
		fmt.Println()
		return
	}

	fmt.Println("✅ Trust line from customer one to cold address created!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	//
	// Send tokens from issuer wallet to customer one wallet
	//
	fmt.Println("⏳ Sending tokens from issuer wallet to customer one wallet...")
	issuerToCustomerOnePayment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(issuerWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(issuerWallet.ClassicAddress),
			Value:    "3800",
		},
		Destination:    types.Address(customerOneWallet.ClassicAddress),
		DestinationTag: 1,
	}

	flattenedTx = issuerToCustomerOnePayment.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = issuerWallet.Sign(flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error signing transaction: %s\n", err)
		return
	}

	response, err = client.SubmitTransactionBlob(txBlob, false)
	if err != nil {
		fmt.Printf("❌ Error submitting transaction: %s\n", err)
		return
	}

	if response.EngineResult != "tesSUCCESS" {
		fmt.Println("❌ Tokens not sent from cold wallet to hot wallet!", response.EngineResult)
		fmt.Println("Try again!")
		fmt.Println()
		return
	}

	fmt.Println("✅ Tokens sent from issuer wallet to customer one wallet!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	// ********************************************************************
	// Send tokens from customer one wallet to customer two
	// ********************************************************************

	fmt.Println("⏳ Sending tokens from issuer wallet to customer two wallet...")
	issuerToCustomerTwoPayment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(issuerWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(issuerWallet.ClassicAddress),
			Value:    "100",
		},
		Destination: types.Address(customerTwoWallet.ClassicAddress),
	}

	flattenedTx = issuerToCustomerTwoPayment.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = issuerWallet.Sign(flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error signing transaction: %s\n", err)
		return
	}

	response, err = client.SubmitTransactionBlob(txBlob, false)
	if err != nil {
		fmt.Printf("❌ Error submitting transaction: %s\n", err)
		return
	}

	if response.EngineResult != "tesSUCCESS" {
		fmt.Println("❌ Tokens not sent from issuer wallet to customer two!", response.EngineResult)
		fmt.Println()
		return
	}

	fmt.Println("✅ Tokens sent from issuer wallet to customer two!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	// ********************************************************************
	// Freeze cold wallet
	// ********************************************************************

	fmt.Println("⏳ Global Freeze...")
	freezeColdWallet := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(issuerWallet.ClassicAddress),
		},
	}

	freezeColdWallet.SetAsfGlobalFreeze()

	flattenedTx = freezeColdWallet.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = issuerWallet.Sign(flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error signing transaction: %s\n", err)
		return
	}

	response, err = client.SubmitTransactionBlob(txBlob, false)
	if err != nil {
		fmt.Printf("❌ Error submitting transaction: %s\n", err)
		return
	}

	if response.EngineResult != "tesSUCCESS" {
		fmt.Println("❌ Cold wallet freezing failed!")
		fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
		fmt.Println()
		return
	}

	fmt.Println("✅ Global Freeze successful!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	// ********************************************************************
	// Try to send tokens from customer one wallet to customer two
	// ********************************************************************

	fmt.Println("⏳ Trying to send tokens from customer one wallet to customer two...")
	customerOneToCustomerTwoPayment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(customerOneWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(issuerWallet.ClassicAddress),
			Value:    "100",
		},
		Destination: types.Address(customerTwoWallet.ClassicAddress),
	}

	flattenedTx = customerOneToCustomerTwoPayment.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = customerOneWallet.Sign(flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error signing transaction: %s\n", err)
		return
	}

	response, err = client.SubmitTransactionBlob(txBlob, false)
	if err != nil {
		fmt.Printf("❌ Error submitting transaction: %s\n", err)
		return
	}

	if response.EngineResult == "tecSUCCESS" {
		fmt.Println("✅ Tokens sent from customer one wallet to customer two!")
		fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
		fmt.Println()
		return
	}

	fmt.Println("❌ Tokens not sent from customer one wallet to customer two!", response.EngineResult)
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	// ********************************************************************
	// Unfreeze global
	// ********************************************************************

	fmt.Println("⏳ Unfreezing global...")
	unfreezeGlobal := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(issuerWallet.ClassicAddress),
		},
	}

	unfreezeGlobal.ClearAsfGlobalFreeze()

	flattenedTx = unfreezeGlobal.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = issuerWallet.Sign(flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error signing transaction: %s\n", err)
		return
	}

	response, err = client.SubmitTransactionBlob(txBlob, false)
	if err != nil {
		fmt.Printf("❌ Error submitting transaction: %s\n", err)
		return
	}

	if response.EngineResult != "tesSUCCESS" {
		fmt.Println("❌ Cold wallet unfreezing failed!")
		fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
		fmt.Println()
		return
	}

	fmt.Println("✅ Cold wallet unfrozen!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	// ********************************************************************
	// Try to send tokens from customer one wallet to customer two
	// ********************************************************************

	fmt.Println("⏳ Trying to send tokens from customer one wallet to customer two...")
	customerOneToCustomerTwoPayment = &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(customerOneWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(issuerWallet.ClassicAddress),
			Value:    "100",
		},
		Destination: types.Address(customerTwoWallet.ClassicAddress),
	}

	flattenedTx = customerOneToCustomerTwoPayment.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = customerOneWallet.Sign(flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error signing transaction: %s\n", err)
		return
	}

	response, err = client.SubmitTransactionBlob(txBlob, false)
	if err != nil {
		fmt.Printf("❌ Error submitting transaction: %s\n", err)
		return
	}

	if response.EngineResult != "tesSUCCESS" {
		fmt.Println("❌ Tokens not sent from customer one wallet to customer two!", response.EngineResult)
		fmt.Println("Try again!")
		return
	}

	fmt.Println("✅ Tokens sent from customer one wallet to customer two!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()
}

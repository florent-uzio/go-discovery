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
	coldWallet, err := xrpl.NewWallet(addresscodec.ED25519)
	if err != nil {
		fmt.Printf("❌ Error creating cold wallet: %s\n", err)
		return
	}
	err = client.FundWallet(&coldWallet)
	if err != nil {
		fmt.Printf("❌ Error funding cold wallet: %s\n", err)
		return
	}
	fmt.Println("💸 Cold wallet funded!")

	hotWallet, err := xrpl.NewWallet(addresscodec.ED25519)
	if err != nil {
		fmt.Printf("❌ Error creating hot wallet: %s\n", err)
		return
	}
	err = client.FundWallet(&hotWallet)
	if err != nil {
		fmt.Printf("❌ Error funding hot wallet: %s\n", err)
		return
	}
	fmt.Println("💸 Hot wallet funded!")

	customerOneWallet, err := xrpl.NewWallet(addresscodec.ED25519)
	if err != nil {
		fmt.Printf("❌ Error creating token wallet: %s\n", err)
		return
	}
	err = client.FundWallet(&customerOneWallet)
	if err != nil {
		fmt.Printf("❌ Error funding customer one wallet: %s\n", err)
		return
	}
	fmt.Println("💸 Customer one wallet funded!")
	fmt.Println()

	fmt.Println("✅ Wallets setup complete!")
	fmt.Println("💳 Cold wallet:", coldWallet.ClassicAddress)
	fmt.Println("💳 Hot wallet:", hotWallet.ClassicAddress)
	fmt.Println("💳 Customer one wallet:", customerOneWallet.ClassicAddress)
	fmt.Println()

	//
	// Configure cold address settings
	//
	fmt.Println("⏳ Configuring cold address settings...")
	coldWalletAccountSet := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(coldWallet.ClassicAddress),
		},
		TickSize:     5,
		TransferRate: 0,
		Domain:       "6578616D706C652E636F6D", // example.com
	}

	coldWalletAccountSet.SetAsfDefaultRipple()
	coldWalletAccountSet.SetDisallowXRP()

	coldWalletAccountSet.SetRequireDestTag()

	flattenedTx := coldWalletAccountSet.Flatten()

	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err := coldWallet.Sign(flattenedTx)
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

	fmt.Println("✅ Cold address settings configured!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	//
	// Configure hot address settings
	//
	fmt.Println("⏳ Configuring hot address settings...")
	hotWalletAccountSet := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(hotWallet.ClassicAddress),
		},
		Domain: "6578616D706C652E636F6D", // example.com
	}

	hotWalletAccountSet.SetAsfRequireAuth()
	hotWalletAccountSet.SetDisallowXRP()
	hotWalletAccountSet.SetRequireDestTag()

	flattenedTx = hotWalletAccountSet.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = hotWallet.Sign(flattenedTx)
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
		fmt.Println("❌ Hot address settings configuration failed!", response.EngineResult)
		fmt.Println("Try again!")
		fmt.Println()
		return
	}

	fmt.Println("✅ Hot address settings configured!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	//
	// Create trust line from hot to cold address
	//
	fmt.Println("⏳ Creating trust line from hot to cold address...")
	hotColdTrustSet := &transactions.TrustSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(hotWallet.ClassicAddress),
		},
		LimitAmount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(coldWallet.ClassicAddress),
			Value:    "100000000000000",
		},
	}

	flattenedTx = hotColdTrustSet.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = hotWallet.Sign(flattenedTx)
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
		fmt.Println("❌ Trust line from hot to cold address creation failed!", response.EngineResult)
		fmt.Println("Try again!")
		fmt.Println()
		return
	}

	fmt.Println("✅ Trust line from hot to cold address created!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	//
	// Create trust line from costumer one to cold address
	//
	fmt.Println("⏳ Creating trust line from customer one to cold address...")
	customerOneColdTrustSet := &transactions.TrustSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(customerOneWallet.ClassicAddress),
		},
		LimitAmount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(coldWallet.ClassicAddress),
			Value:    "100000000000000",
		},
	}

	flattenedTx = customerOneColdTrustSet.Flatten()
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
		fmt.Println("❌ Trust line from customer one to cold address creation failed!", response.EngineResult)
		fmt.Println("Try again!")
		fmt.Println()
		return
	}

	fmt.Println("✅ Trust line from customer one to cold address created!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	//
	// Send tokens from cold wallet to hot wallet
	//
	fmt.Println("⏳ Sending tokens from cold wallet to hot wallet...")
	coldToHotPayment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(coldWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(coldWallet.ClassicAddress),
			Value:    "3800",
		},
		Destination:    types.Address(hotWallet.ClassicAddress),
		DestinationTag: 1,
	}

	flattenedTx = coldToHotPayment.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = coldWallet.Sign(flattenedTx)
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

	fmt.Println("✅ Tokens sent from cold wallet to hot wallet!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	//
	// Send tokens from hot wallet to customer one
	//
	fmt.Println("⏳ Sending tokens from cold wallet to customer one...")
	coldToCustomerOnePayment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(coldWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(coldWallet.ClassicAddress),
			Value:    "100",
		},
		Destination: types.Address(customerOneWallet.ClassicAddress),
	}

	flattenedTx = coldToCustomerOnePayment.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = coldWallet.Sign(flattenedTx)
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
		fmt.Println("❌ Tokens not sent from cold wallet to customer one!", response.EngineResult)
		fmt.Println()
		return
	}

	fmt.Println("✅ Tokens sent from cold wallet to customer one!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	//
	// Freeze cold wallet
	//
	fmt.Println("⏳ Freezing cold wallet...")
	freezeColdWallet := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(coldWallet.ClassicAddress),
		},
	}

	freezeColdWallet.SetAsfGlobalFreeze()

	flattenedTx = freezeColdWallet.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = coldWallet.Sign(flattenedTx)
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

	fmt.Println("✅ Cold wallet frozen!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	//
	// Try to send tokens from hot wallet to customer one
	//
	fmt.Println("⏳ Trying to send tokens from hot wallet to customer one...")
	hotToCustomerOnePayment := &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(hotWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(coldWallet.ClassicAddress),
			Value:    "100",
		},
		Destination: types.Address(customerOneWallet.ClassicAddress),
	}

	flattenedTx = hotToCustomerOnePayment.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = hotWallet.Sign(flattenedTx)
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
		fmt.Println("✅ Tokens sent from hot wallet to customer one!")
		fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
		fmt.Println()
		return
	}

	fmt.Println("❌ Tokens not sent from hot wallet to customer one!", response.EngineResult)
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()

	// //
	// // Unfreeze cold wallet
	// //
	fmt.Println("⏳ Unfreezing cold wallet...")
	unfreezeColdWallet := &transactions.AccountSet{
		BaseTx: transactions.BaseTx{
			Account: types.Address(coldWallet.ClassicAddress),
		},
	}

	unfreezeColdWallet.ClearAsfGlobalFreeze()

	flattenedTx = unfreezeColdWallet.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = coldWallet.Sign(flattenedTx)
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

	//
	// Try to send tokens from hot wallet to customer one
	//
	fmt.Println("⏳ Trying to send tokens from hot wallet to customer one...")
	hotToCustomerOnePayment = &transactions.Payment{
		BaseTx: transactions.BaseTx{
			Account: types.Address(hotWallet.ClassicAddress),
		},
		Amount: types.IssuedCurrencyAmount{
			Currency: currencyCode,
			Issuer:   types.Address(coldWallet.ClassicAddress),
			Value:    "100",
		},
		Destination: types.Address(customerOneWallet.ClassicAddress),
	}

	flattenedTx = hotToCustomerOnePayment.Flatten()
	err = client.Autofill(&flattenedTx)
	if err != nil {
		fmt.Printf("❌ Error autofilling transaction: %s\n", err)
		return
	}

	txBlob, _, err = hotWallet.Sign(flattenedTx)
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
		fmt.Println("❌ Tokens not sent from hot wallet to customer one!", response.EngineResult)
		fmt.Println("Try again!")
		return
	}

	fmt.Println("✅ Tokens sent from hot wallet to customer one!")
	fmt.Printf("🌐 Hash: %s\n", response.Tx["hash"])
	fmt.Println()
}

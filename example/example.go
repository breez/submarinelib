package main

import (
	"fmt"
	submarine "github.com/breez/submarinelib"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil"
)

func main() {
	fmt.Println("Hello submarine.")
	// Generate two secrets
	lightningPublicKey, lightningPrivateKey := submarine.GenSecret()
	payeeChainPublicKey, payeeChainPrivateKey, _ := submarine.GenPublicPrivateKeypair()
	payerChainPublicKey, payerChainPrivateKey, _ := submarine.GenPublicPrivateKeypair()

	fmt.Println(lightningPublicKey)
	fmt.Println(btcutil.Hash160(lightningPublicKey))

	fmt.Println(lightningPrivateKey)
	fmt.Println(payeeChainPublicKey)
	fmt.Println(payeeChainPrivateKey)
	fmt.Println(payerChainPublicKey)
	fmt.Println(payerChainPrivateKey)

	// Create a script with our and client's data
	script, _ := submarine.GenSubmarineSwapScript(payeeChainPublicKey, payerChainPublicKey, lightningPublicKey, 600)
	fmt.Println(script)

	var network *chaincfg.Params
	network = &chaincfg.SimNetParams

	// Now it's time to create a nice address
	address := submarine.GenBase58Address(script, network)
	fmt.Println(address)
}
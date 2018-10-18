package submarinelib

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/btcsuite/btcutil"
)

func GenSubmarineSwapScript(payeePubKey, payerPubKey, preimageHash []byte, lockHeight int64) ([]byte, error) {
	builder := txscript.NewScriptBuilder()

	builder.AddOp(txscript.OP_HASH160)
	builder.AddData(btcutil.Hash160(preimageHash))
	builder.AddOp(txscript.OP_EQUAL) // Leaves 0P1 (true) on the stack if preimage matches
	builder.AddOp(txscript.OP_IF)
	builder.AddData(payeePubKey) // Path taken if preimage matches
	builder.AddOp(txscript.OP_ELSE)
	builder.AddInt64(lockHeight)
	builder.AddOp(txscript.OP_CHECKSEQUENCEVERIFY)
	builder.AddOp(txscript.OP_DROP)
	builder.AddData(payerPubKey) // Refund back to payer
	builder.AddOp(txscript.OP_ENDIF)
	builder.AddOp(txscript.OP_CHECKSIG)

	return builder.Script()
}

func GenSecret() ([]byte, []byte) {
	var secret [32]byte
	rand.Read(secret[:])
	secretHash := sha256.Sum256(secret[:])
	return secretHash[:], secret[:]
}

func GenPublicPrivateKeypair () ([]byte, []byte, error) {
	key, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, nil, err
	}
	return key.PubKey().SerializeCompressed(), key.Serialize(), nil
}

func GenBase58Address(serializedScript []byte, net *chaincfg.Params) string {
	scriptHash, _ := btcutil.NewAddressScriptHash(serializedScript, net)
	return scriptHash.String()
}

func GetRedeemTransaction(totalAmount int64, fee int64, swapTransaction [32]byte, serializedScript []byte, privateKeyBytes []byte, preimage []byte, redeemAddress btcutil.Address) (string, error) {
	// Redeem as much as possible, after substracting the fee
	redeemAmount := totalAmount - fee

	// Type 2 supports CSV
	redeemTx := wire.NewMsgTx(2)

	// We need to reference the swap transactions outpoint
	var hash chainhash.Hash = swapTransaction
	prevOut := wire.NewOutPoint(&hash, 0)


	// Send the funds to an address
	redeemScript, err := txscript.PayToAddrScript(redeemAddress)
	if err != nil {
		return "", err
	}

	txOut := wire.NewTxOut(redeemAmount, redeemScript)
	redeemTx.AddTxOut(txOut)

	// Sign with out private key
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)
	scriptSig, err := txscript.SignatureScript(redeemTx, 0, serializedScript, txscript.SigHashAll, privateKey, true)
	if err != nil {
		return "", err
	}

	txIn := wire.NewTxIn(prevOut, serializedScript, [][]byte{scriptSig, preimage})
	redeemTx.AddTxIn(txIn)

	// Serialize the transaction and convert to hex string.
	buf := bytes.NewBuffer(make([]byte, 0, redeemTx.SerializeSize()))
	err = redeemTx.Serialize(buf)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf.Bytes()), nil
}

func GetRefundTransaction(totalAmount int64, fee int64, swapTransaction [32]byte, serializedScript []byte, privateKeyBytes []byte, refundAddress btcutil.Address) (string, error) {
	// Refund as much as possible, after substracting the fee
	redeemAmount := totalAmount - fee

	// Type 2 supports CSV
	redeemTx := wire.NewMsgTx(2)

	// We need to reference the swap transactions outpoint
	var hash chainhash.Hash = swapTransaction
	prevOut := wire.NewOutPoint(&hash, 0)


	// Send the funds to the refund address
	redeemScript, err := txscript.PayToAddrScript(refundAddress)
	if err != nil {
		return "", err
	}

	txOut := wire.NewTxOut(redeemAmount, redeemScript)
	redeemTx.AddTxOut(txOut)

	// Sign with out private key
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyBytes)
	scriptSig, err := txscript.SignatureScript(redeemTx, 0, serializedScript, txscript.SigHashAll, privateKey, true)
	if err != nil {
		return "", err
	}

	txIn := wire.NewTxIn(prevOut, serializedScript, [][]byte{scriptSig, {txscript.OP_0}})
	redeemTx.AddTxIn(txIn)

	// Serialize the transaction and convert to hex string.
	buf := bytes.NewBuffer(make([]byte, 0, redeemTx.SerializeSize()))
	err = redeemTx.Serialize(buf)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(buf.Bytes()), nil
}
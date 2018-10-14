package submarinelib

import (
	"crypto/rand"
	"crypto/sha256"
	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
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
	builder.AddOp(txscript.OP_CHECKLOCKTIMEVERIFY) // Script will fail here if lockheight higher than current block
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
	sessionKey, err := btcec.NewPrivateKey(btcec.S256())
	if err != nil {
		return nil, nil, err
	}
	return sessionKey.PubKey().SerializeCompressed(), sessionKey.Serialize(), nil
}

func GenBase58Address(serializedScript []byte, net *chaincfg.Params) string {
	scriptHash, _ := btcutil.NewAddressScriptHash(serializedScript, net)
	return scriptHash.String()
}
package submarinelib

import (
	"crypto/rand"
	"crypto/sha256"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcutil"
)

func GenSubmarineSwapScript(aPubKeyHash, bPubKeyHash, preimageHash []byte, lockHeight int64) ([]byte, error) {
	builder := txscript.NewScriptBuilder()

	builder.AddOp(txscript.OP_HASH160)
	builder.AddData(btcutil.Hash160(preimageHash))
	builder.AddOp(txscript.OP_EQUAL) // Leaves 0P1 (true) on the stack if preimage matches
	builder.AddOp(txscript.OP_IF)
	builder.AddData(aPubKeyHash) // Path taken if preimage matches
	builder.AddOp(txscript.OP_ELSE)
	builder.AddInt64(lockHeight)
	builder.AddOp(txscript.OP_CHECKLOCKTIMEVERIFY) // Script will fail here if lockheight higher than current block
	builder.AddOp(txscript.OP_DROP)
	builder.AddData(bPubKeyHash)
	builder.AddOp(txscript.OP_ENDIF)
	builder.AddOp(txscript.OP_CHECKSIG)

	return builder.Script()
}

func GenPreimage () ([32]byte, [32]byte) {
	var preimage [32]byte
	rand.Read(preimage[:])
	rHash := sha256.Sum256(preimage[:])
	return rHash, preimage
}




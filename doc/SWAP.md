Credit for inventing submarine swaps goes to [Alex Bosworth](https://github.com/alexbosworth). This project draws *heavily* from his [work](https://github.com/submarineswaps/swaps-service), to the point of it simply being a Golang implmenentation. The major motivation was enabling users of [Breez](https://github.com/breez/breezmobile) to top-up their lighning network channels trustlessly.

## Use case

### Alice has a lightning network channel open with the Breez (or Bob) node. Her side of the channel is low on funds so she decides to use her on-chain funds to top up:

1. Via an API provided by Bob, whom she has an off-chain channel with, Alice requests an address to pay into. These funds will be used to top up her channel and payment can be independent of Alice's wallet software (can be from an exchange or similar)
2. Alice generates a private/public keypair and a preimage and then provides Bob with an LN payment hash (the preimage hashed twice, once with SHA-256 and once with RIPEMD-160, resulting in 20 bytes) and the pubkey (33 bytes) for refund in case anything goes wrong. The secrets used to generate these hashes she keeps to herself
3. Bob takes that information from Alice, generates his own private/public keypair (used in case the swap is succesful), and proceeds to use all that data to craft a bitcoin script as follows:

```xml
OP_HASH160 <paymentHash> OP_EQUAL
OP_IF
  <bobPubKey>
OP_ELSE
  <currentBlockHeight + 72> OP_CHECKLOCKTIMEVERIFY OP_DROP <alicePubkey>
OP_ENDIF
OP_CHECKSIG
```

4. Bob generates a Base58Check formatted address from the script above
5. Bob sends the address and his pubkey back to Alice
6. Alice uses Bob's pubkey along with her payment hash and pubkey to create an exact same transaction Bob previously created
7. Just like Bob, Alice generates a Base58Check address from the script and checks the two addresses for equality
8. If the two addresses match, Alice now pays to that address
9. Bob now sees (on the blockchain) that Alice has indeed paid to this address and, via an API, asks her for an LN invoice of that amount
10. Alice creates an LN invoice with the payment hash and its corresponding preimage secret created in step 2
11. Alice sends the bolt11 invoice to Bab and now he has until lockheight to fullfill his obligation
12. Bob pays Alice's invoices and is rewarded with the preimage secret throught the lightning network
13. With this proof-of-payment Bob creates a transaction to reedeem the funds, the transaction is equipped with his signature and the 32 byte preimage to unlock the funds thusly:

```xml
<bobSignature> <preimage>
OP_HASH160 <paymentHash> OP_EQUAL // after <preimage> is OP_HASH160'd the OP_EQUAL evaluates to true 
OP_IF
  <bobPubKey> // We are left with only Bob's pubkey on the stack 
OP_ELSE
  // This path is not taken
OP_ENDIF
OP_CHECKSIG
```

14. With only `<bobPubKey>` on the stack, check against `<bobSignature>` returns true and with the private key from step 3 Bob is now in control of the funds
15. He transfers them elsewhere, Alice got his off-chain and Bob got the on-chain payment

### In case Bob reneges on his obligation to pay Alice's off-chain invoice:

16. Alice needs to wait until lockheight (72 blocks or about 12 hours from when her and Bob agreed on the swap)
17. Now that the locktime has passed Alice broadcasts a transaction of her own

```xml
<aliceSignature> OP_0
OP_HASH160 <paymentHash> OP_EQUAL // <paymentHash> doesn't match OP_HASH160'd OP_0
OP_IF
  // So this path is not taken
OP_ELSE
  <cltvExpiry> OP_CHECKLOCKTIMEVERIFY OP_DROP // We check if the lockheight had passed
  <alicePubKey> // <alicePubKey> is pushed onto the stack
OP_ENDIF
OP_CHECKSIG 
```

18. With only `<alicePubKey>` on the stack, check against `<aliceSignature>` returns true and Alice is now in control of the funds
19. Alice has learned her lesson and will not be conducting business with Bob in the future

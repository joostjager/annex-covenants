# Annex covenant demo

This repository is intended to be complementary to the [mailing list post](https://lists.linuxfoundation.org/pipermail/bitcoin-dev/2023-June/021758.html) in which the potential of the taproot annex for enabling presigned transaction-based covenants is covered.

While the post describes how time-locked vaults using presigned transactions can benefit from the taproot annex, this repository only shows an implementation of the simplest possible covenant based on a presigned transaction. My hope is that no unreasonable amount of imagination is required to extrapolate this to more advanced applications.

## Goal

The goal of this demo is to show how coins can be locked into a covenant that only allows spending to a predefined destination address `bcrt1p2uu43ca9hzyjqjtvjl0xehx47rjj8szclsc2kg98utfq678z7n8qftt3gh` at a later time. Of course you can argue that there is no use for this mechanism if there is no other spend path anyway. But as mentioned before, I just want to demonstrate the concept.

`WALLET -> COVENANT -> PRE-DEFINED DESTINATION`

## Prerequisites

* For the wallet, a Bitcoin Core can be used to which a [patch](https://github.com/joostjager/bitcoin/tree/psbt-annex) is applied that allows signing psbts containing an annex.

  Additionally the instance needs to run with `acceptnonstdtxn=1` to accept annex-containing transactions.

* Clone this [`annex-covenant` repository](https://github.com/joostjager/annex-covenant)

* Build the `annex-covenant` binary by running `go build .`


## Create

### 1. Fund a psbt using Bitcoin Core

`bitcoin-cli walletcreatefundedpsbt '[]' '[{"bcrt1pqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqm3usuw": 0.01}]' 0 '{"changePosition": 1}'`

```json
{
"psbt": "cHNidP8BAIkCAAAAATLkCDs6GGbkOG02iH3xCSsqO86C9cvzW1IYqLlM3cr/AQAAAAD9////AkBCDwAAAAAAIlEgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABI5ptJAAAAACJRID62TtsjhjPdOK5IitMCFc5IDCpefHA9Tx3HpZRqrchrAAAAAAABASucRqtJAAAAACJRIBJoMOW68yCXa1Hxl8f56eLUsgv96s1tMkYBrHnu16RQIRbcESnFKNjDXHzXKkwaJLKod2XGF4lLE91j/81T5FqfCRkAZ9hbDFYAAIABAACAAAAAgAEAAAAKAAAAARcg3BEpxSjYw1x81ypMGiSyqHdlxheJSxPdY//NU+RanwkAAAEFIAYZj4TrbyXvbA+4quCciBBwD2+Pb6FME7WebczHV86eIQcGGY+E628l72wPuKrgnIgQcA9vj2+hTBO1nm3Mx1fOnhkAZ9hbDFYAAIABAACAAAAAgAEAAAALAAAAAA==",
"fee": 0.00007700,
"changepos": 1
}
```

The destination is a dummy address that will later be replaced by the covenant address. The change position is fixed so that the covenant output can always be assumed to be 0.

### 2. Update the psbt from above so that it creates the covenant

`./annex-covenant create <psbt>`

This returns the updated psbt:

```
cHNidP8BAIkCAAAAATLkCDs6GGbkOG02iH3xCSsqO86C9cvzW1IYqLlM3cr/AQAAAAD9////AkBCDwAAAAAAIlEgX9V9DK0E10+5lWgfou0ECi2/0H0wAUnxk2spaTLXUZVI5ptJAAAAACJRID62TtsjhjPdOK5IitMCFc5IDCpefHA9Tx3HpZRqrchrAAAAAAABASucRqtJAAAAACJRIBJoMOW68yCXa1Hxl8f56eLUsgv96s1tMkYBrHnu16RQIRbcESnFKNjDXHzXKkwaJLKod2XGF4lLE91j/81T5FqfCRkAZ9hbDFYAAIABAACAAAAAgAEAAAAKAAAAARcg3BEpxSjYw1x81ypMGiSyqHdlxheJSxPdY//NU+RanwkFYW5uZXhAqPTS4EvezTsQzERFy0MOS/isLK7kxe14pROQH7tkan+X76/XlReukZTqQBZ8jj9jHpz/FHbZZjR6nROvShqcUAAAAQUgBhmPhOtvJe9sD7iq4JyIEHAPb49voUwTtZ5tzMdXzp4hBwYZj4TrbyXvbA+4quCciBBwD2+Pb6FME7WebczHV86eGQBn2FsMVgAAgAEAAIAAAACAAQAAAAsAAAAA
```

In decoded form the following changes become apparent:
* The dummy address `bcrt1pqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqm3usuw` of output 0 has been replaced by the covenant address.
* Input 0 has been decorated with an unknown tag with the key `annex` and as the value the ephemeral signature of the spend transaction.

```json
{
        "tx": {
                "txid": "7cf2e70aed40f2009db37634050f324a600d7fb9eda80f2818fdd4f38f9d88d5",
                "hash": "7cf2e70aed40f2009db37634050f324a600d7fb9eda80f2818fdd4f38f9d88d5",
                "version": 2,
                "size": 137,
                "vsize": 137,
                "weight": 548,
                "locktime": 0,
                "vin": [
                        {
                                "txid": "ffcadd4cb9a818525bf3cbf582ce3b2a2b09f17d88366d38e466183a3b08e432",
                                "vout": 1,
                                "scriptSig": {
                                        "asm": "",
                                        "hex": ""
                                },
                                "sequence": 4294967293
                        }
                ],
                "vout": [
                        {
                                "value": 0.01000000,
                                "n": 0,
                                "scriptPubKey": {
                                        "asm": "1 5fd57d0cad04d74fb995681fa2ed040a2dbfd07d300149f1936b296932d75195",
                                        "desc": "rawtr(5fd57d0cad04d74fb995681fa2ed040a2dbfd07d300149f1936b296932d75195)#9p0f9c8z",
                                        "hex": "51205fd57d0cad04d74fb995681fa2ed040a2dbfd07d300149f1936b296932d75195",
                                        "address": "bcrt1ptl2h6r9dqnt5lwv4dq069mgypgkml5raxqq5nuvndv5kjvkh2x2spk8j3t",
                                        "type": "witness_v1_taproot"
                                }
                        },
                        {
                                "value": 12.34953800,
                                "n": 1,
                                "scriptPubKey": {
                                        "asm": "1 3eb64edb238633dd38ae488ad30215ce480c2a5e7c703d4f1dc7a5946aadc86b",
                                        "desc": "rawtr(3eb64edb238633dd38ae488ad30215ce480c2a5e7c703d4f1dc7a5946aadc86b)#fel0qy2f",
                                        "hex": "51203eb64edb238633dd38ae488ad30215ce480c2a5e7c703d4f1dc7a5946aadc86b",
                                        "address": "bcrt1p86myakerscea6w9wfz9dxqs4eeyqc2j703cr6ncac7jeg64dep4s4r7xfm",
                                        "type": "witness_v1_taproot"
                                }
                        }
                ]
        },
        "global_xpubs": [],
        "psbt_version": 0,
        "proprietary": [],
        "unknown": {},
        "inputs": [
                {
                        "witness_utxo": {
                                "amount": 12.35961500,
                                "scriptPubKey": {
                                        "asm": "1 126830e5baf320976b51f197c7f9e9e2d4b20bfdeacd6d324601ac79eed7a450",
                                        "desc": "rawtr(126830e5baf320976b51f197c7f9e9e2d4b20bfdeacd6d324601ac79eed7a450)#x6k4yhud",
                                        "hex": "5120126830e5baf320976b51f197c7f9e9e2d4b20bfdeacd6d324601ac79eed7a450",
                                        "address": "bcrt1pzf5rped67vsfw6637xtu070fut2tyzlaatxk6vjxqxk8nmkh53gqqj4e3m",
                                        "type": "witness_v1_taproot"
                                }
                        },
                        "taproot_bip32_derivs": [
                                {
                                        "pubkey": "dc1129c528d8c35c7cd72a4c1a24b2a87765c617894b13dd63ffcd53e45a9f09",
                                        "master_fingerprint": "67d85b0c",
                                        "path": "m/86h/1h/0h/1/10",
                                        "leaf_hashes": []
                                }
                        ],
                        "taproot_internal_key": "dc1129c528d8c35c7cd72a4c1a24b2a87765c617894b13dd63ffcd53e45a9f09",
                        "unknown": {
                                "616e6e6578": "a8f4d2e04bdecd3b10cc4445cb430e4bf8ac2caee4c5ed78a513901fbb646a7f97efafd79517ae9194ea40167c8e3f631e9cff1476d966347a9d13af4a1a9c50"
                        }
                }
        ],
        "outputs": [
                {},
                {
                        "taproot_internal_key": "06198f84eb6f25ef6c0fb8aae09c8810700f6f8f6fa14c13b59e6dccc757ce9e",
                        "taproot_bip32_derivs": [
                                {
                                        "pubkey": "06198f84eb6f25ef6c0fb8aae09c8810700f6f8f6fa14c13b59e6dccc757ce9e",
                                        "master_fingerprint": "67d85b0c",
                                        "path": "m/86h/1h/0h/1/11",
                                        "leaf_hashes": []
                                }
                        ]
                }
        ],
        "fee": 0.00007700
}
```

### 3. Sign, finalize and send using Bitcoin Core

`bitcoin-cli walletprocesspsbt <updated psbt>`

`bitcoin-cli finalizepsbt <signed psbt>`

This returns the covenant tx:

```json
{
"hex": "0200000000010132e4083b3a1866e4386d36887df1092b2a3bce82f5cbf35b5218a8b94cddcaff0100000000fdffffff0240420f00000000002251205fd57d0cad04d74fb995681fa2ed040a2dbfd07d300149f1936b296932d7519548e69b49000000002251203eb64edb238633dd38ae488ad30215ce480c2a5e7c703d4f1dc7a5946aadc86b0240b928550174b24f00982b4fa8839d66a72c5e652867fb1eb80c6c7d7b7852fa5801a27f88229ce336322d50a21176115bf1cbb27586a52a3405600fafbd0d581c4150a8f4d2e04bdecd3b10cc4445cb430e4bf8ac2caee4c5ed78a513901fbb646a7f97efafd79517ae9194ea40167c8e3f631e9cff1476d966347a9d13af4a1a9c5000000000",
"complete": true
}
```

Then send it using:

`bitcoin-cli sendrawtransaction <covenant tx hex>`

The covenant has now been created.

## Spend

### 4. Reconstruct the spend transaction

Use the raw covenant transaction as published in the create step. Knowing the txid of it is enough, because the full transaction can be retrieved from the chain. There is no additional state required, allowing the tool `annex-covenant` to be stateless.

`./annex-covenant spend <covenant tx hex>`

This returns a complete spend transaction that is ready for broadcast:

```
01000000000101d5889d8ff3d4fd18280fa8edb97f0d604a320f053476b39d00f240ed0ae7f27c00000000000000000001703a0f0000000000225120573958e3a5b88920496c97de6cdcd5f0e523c058fc30ab20a7e2d20d78e2f4ce0140a8f4d2e04bdecd3b10cc4445cb430e4bf8ac2caee4c5ed78a513901fbb646a7f97efafd79517ae9194ea40167c8e3f631e9cff1476d966347a9d13af4a1a9c5000000000
```

In decoded form it is visible that it spends to the pre-defined address `bcrt1p2uu43ca9hzyjqjtvjl0xehx47rjj8szclsc2kg98utfq678z7n8qftt3gh`. Because the private key used to create this transaction has been discarded after the create step, it isn't possible to spend the funds anywhere else.

```json
{
        "txid": "88e687d69cf1bad5fc00cd56a6f73e3614cdde805933c2bdde66df79ba0c8863",
        "hash": "3c108b5317441bf80d102a265996f76682f9886a62e28556b2d4e27d21aa64df",
        "version": 1,
        "size": 162,
        "vsize": 111,
        "weight": 444,
        "locktime": 0,
        "vin": [
                {
                        "txid": "7cf2e70aed40f2009db37634050f324a600d7fb9eda80f2818fdd4f38f9d88d5",
                        "vout": 0,
                        "scriptSig": {
                                "asm": "",
                                "hex": ""
                        },
                        "txinwitness": [
                                "a8f4d2e04bdecd3b10cc4445cb430e4bf8ac2caee4c5ed78a513901fbb646a7f97efafd79517ae9194ea40167c8e3f631e9cff1476d966347a9d13af4a1a9c50"
                        ],
                        "sequence": 0
                }
        ],
        "vout": [
                {
                        "value": 0.00998000,
                        "n": 0,
                        "scriptPubKey": {
                                "asm": "1 573958e3a5b88920496c97de6cdcd5f0e523c058fc30ab20a7e2d20d78e2f4ce",
                                "desc": "rawtr(573958e3a5b88920496c97de6cdcd5f0e523c058fc30ab20a7e2d20d78e2f4ce)#csukkllk",
                                "hex": "5120573958e3a5b88920496c97de6cdcd5f0e523c058fc30ab20a7e2d20d78e2f4ce",
                                "address": "bcrt1p2uu43ca9hzyjqjtvjl0xehx47rjj8szclsc2kg98utfq678z7n8qftt3gh",
                                "type": "witness_v1_taproot"
                        }
                }
        ]
}
```

### 5.  Then send it using

`bitcoin-cli sendrawtransaction <spend tx hex>`

The covenant has now been spent.

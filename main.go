package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/schnorr"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/psbt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
	"github.com/urfave/cli/v2"
)

/*
Steps:

1. Fund psbt paying to dummy address.

bitcoin-cli walletcreatefundedpsbt '[]' '[{"bcrt1pqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqm3usuw": 0.01}]' 0 '{"changePosition": 1}'

2. Run `create`` to update the psbt with the covenant.
3. Sign and broadcast psbt.

To spend:

4. Run `spend` to create the spend transaction.
5. Broadcast spend transaction.
*/

const (
	destAddress         = "bcrt1p2uu43ca9hzyjqjtvjl0xehx47rjj8szclsc2kg98utfq678z7n8qftt3gh"
	fixedAbsoluteFeeSat = 2000
	fixedOutputIndex    = 0
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:      "create",
				Usage:     "Create a covenant transaction spending to a fixed destination address",
				ArgsUsage: "<psbt>",
				Action: func(cCtx *cli.Context) error {
					if cCtx.Args().Len() != 1 {
						return cli.ShowSubcommandHelp(cCtx)
					}

					return create(cCtx.Args().First())
				},
			},
			{
				Name:      "spend",
				Usage:     "Spend a covenant transaction to a pre-defined destination address",
				ArgsUsage: "<hex-encoded tx>",
				Action: func(cCtx *cli.Context) error {
					if cCtx.Args().Len() != 1 {
						return cli.ShowSubcommandHelp(cCtx)
					}

					return spend(cCtx.Args().First())
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func spend(txHex string) error {
	// Decode raw transaction.
	tx, err := hex.DecodeString(txHex)
	if err != nil {
		return err
	}

	var covenantTx wire.MsgTx
	err = covenantTx.Deserialize(bytes.NewReader(tx))
	if err != nil {
		return fmt.Errorf("error deserializing raw covenant tx: %w", err)
	}

	// Extract signature from the annex.
	sig := covenantTx.TxIn[0].Witness[1][1:]
	if len(sig) != 64 {
		return fmt.Errorf("invalid signature length %v", len(sig))
	}

	// Reconstruct the presigned spend transaction.
	spendTx := wire.NewMsgTx(wire.TxVersion)
	spendTx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			Hash:  covenantTx.TxHash(),
			Index: uint32(fixedOutputIndex),
		},
	})

	destAddr, err := btcutil.DecodeAddress(destAddress, &chaincfg.RegressionNetParams)
	if err != nil {
		return err
	}
	destPkScript, err := txscript.PayToAddrScript(destAddr)
	if err != nil {
		return err
	}

	spendTx.AddTxOut(&wire.TxOut{
		PkScript: destPkScript,
		Value:    covenantTx.TxOut[fixedOutputIndex].Value - fixedAbsoluteFeeSat,
	})

	// Insert extract signature to complete the spend transaction.
	spendTx.TxIn[0].Witness = wire.TxWitness{sig}

	// Output the spend transaction in hex format.
	var presignedBuffer bytes.Buffer
	err = spendTx.Serialize(&presignedBuffer)
	if err != nil {
		return err
	}

	fmt.Println(hex.EncodeToString(presignedBuffer.Bytes()))

	return nil
}

func create(inputPsbt string) error {
	// Decode psbt.
	r := strings.NewReader(inputPsbt)
	packet, err := psbt.NewFromRawBytes(r, true)
	if err != nil {
		return err
	}

	// Generate an ephemeral key pair for the covenant.
	privTaprootKey, err := btcec.NewPrivateKey()
	if err != nil {
		return err
	}

	taprootKey := privTaprootKey.PubKey()

	tapScriptAddr, err := btcutil.NewAddressTaproot(
		schnorr.SerializePubKey(taprootKey), &chaincfg.RegressionNetParams,
	)
	if err != nil {
		return err
	}

	// Calculate the pkscript for inserting into the psbt.
	pkScript, err := txscript.PayToAddrScript(tapScriptAddr)
	if err != nil {
		return err
	}

	// Update the psbt dummy output with the covenant pkscript.
	packet.Outputs[fixedOutputIndex] = psbt.POutput{}
	packet.UnsignedTx.TxOut[fixedOutputIndex].PkScript = pkScript

	// The txid of the covenant transaction is now final.
	covenantTxHash := packet.UnsignedTx.TxHash()

	// Construct the spending transaction.
	spendingTx := wire.NewMsgTx(wire.TxVersion)

	// Spend from the covenant transaction which now has a final txid.
	spendingTx.AddTxIn(&wire.TxIn{
		PreviousOutPoint: wire.OutPoint{
			Hash:  covenantTxHash,
			Index: uint32(fixedOutputIndex),
		},
	})

	// Spend to the predefined destination address. Use a fixed absoluite
	// fee.
	destAddr, err := btcutil.DecodeAddress(destAddress, &chaincfg.RegressionNetParams)
	if err != nil {
		return err
	}
	destPkScript, err := txscript.PayToAddrScript(destAddr)
	if err != nil {
		return err
	}

	spendingTx.AddTxOut(&wire.TxOut{
		PkScript: destPkScript,
		Value:    packet.UnsignedTx.TxOut[fixedOutputIndex].Value - fixedAbsoluteFeeSat,
	})

	// Sign the spend transaction using the ephemeral key.
	fetcher := txscript.NewCannedPrevOutputFetcher(
		packet.UnsignedTx.TxOut[fixedOutputIndex].PkScript,
		packet.UnsignedTx.TxOut[fixedOutputIndex].Value,
	)
	sigHashes := txscript.NewTxSigHashes(spendingTx, fetcher)
	sigHash, err := txscript.CalcTaprootSignatureHash(
		sigHashes,
		txscript.SigHashDefault,
		spendingTx,
		0,
		fetcher,
	)
	if err != nil {
		return err
	}
	sig, err := schnorr.Sign(privTaprootKey, sigHash)
	if err != nil {
		return err
	}

	// Store the ephemeral signer signature in the annex of the covenant
	// transaction. Losing the signature would mean that the covenant output
	// is unspendable. By storing it in the annex, we can always recover the
	// signature and spend the covenant output. This is assuming that there
	// will always be bitcoin nodes around that do not prune witness data.
	packet.Inputs[0].Unknowns = []*psbt.Unknown{
		{
			Key:   []byte("annex"),
			Value: sig.Serialize(),
		},
	}

	// Output the updated psbt in base64 format.
	var b bytes.Buffer
	err = packet.Serialize(&b)
	if err != nil {
		return err
	}

	encodedPsbt := base64.StdEncoding.EncodeToString(b.Bytes())
	fmt.Println(encodedPsbt)

	return nil
}

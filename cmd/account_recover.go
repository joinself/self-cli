// Copyright 2020 Self Group Ltd. All Rights Reserved.

package cmd

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joinself/self-go-sdk/pkg/ntp"
	"github.com/joinself/self-go-sdk/pkg/siggraph"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
)

var accountRecoverCommand = &cobra.Command{
	Use:   "recover",
	Short: "recover an account with a recovery key",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			check(errors.New("you must specify an app identity and device [appID]"))
		}

		if recoveryKey == "" {
			check(errors.New("You must provide a secret recovery key"))
		}

		if strings.Contains(recoveryKey, "_") {
			keyParts := strings.Split(recoveryKey, "_")
			if keyParts[0] != "rk_" {
				check(errors.New("the recovery key provided is not valid, it should start with 'rk'"))
			}
			recoveryKey = keyParts[1]
		}

		var edpk, edsk, erpk, ersk string

		if devicePublicKey != "" {
			edpk = devicePublicKey
		} else {
			dpk, dsk, err := ed25519.GenerateKey(rand.Reader)
			check(err)

			edpk = enc.EncodeToString(dpk)
			edsk = base64.RawStdEncoding.EncodeToString(dsk.Seed())
		}

		if recoveryPublicKey != "" {
			erpk = recoveryPublicKey
		} else {
			rpk, rsk, err := ed25519.GenerateKey(rand.Reader)
			check(err)

			erpk = enc.EncodeToString(rpk)
			ersk = base64.RawStdEncoding.EncodeToString(rsk.Seed())
		}

		client := rest(args[0], recoveryKey)

		done := make(chan error)

		// get the identity history
		go log("getting identity history", done)

		resp, err := client.Get("/v1/identities/" + args[0] + "/history")
		done <- err

		if err != nil {
			os.Exit(1)
		}

		// load the signature graph
		var history []json.RawMessage

		err = json.Unmarshal(resp, &history)
		check(err)

		sg, err := siggraph.New(history)
		check(err)

		// create a new operation
		rkid := strconv.Itoa(len(sg.Keys()) + 1)
		dkid := strconv.Itoa(len(sg.Keys()) + 2)
		ddid := strconv.Itoa(len(sg.Devices()) + 1)
		now := ntp.TimeFunc().Unix()

		if effectiveFrom == 0 {
			effectiveFrom = int(now)
		}

		actions := []siggraph.Action{
			{
				KID:           strings.Split(recoveryKey, ":")[0],
				Type:          siggraph.TypeRecoveryKey,
				Action:        siggraph.ActionKeyRevoke,
				EffectiveFrom: int64(effectiveFrom),
			},
			{
				KID:           dkid,
				DID:           ddid,
				Type:          siggraph.TypeDeviceKey,
				Action:        siggraph.ActionKeyAdd,
				EffectiveFrom: now,
				Key:           edpk,
			},
			{
				KID:           rkid,
				Type:          siggraph.TypeRecoveryKey,
				Action:        siggraph.ActionKeyAdd,
				EffectiveFrom: now,
				Key:           erpk,
			},
		}

		operation := newOperation(sg, actions, recoveryKey)

		// check the operation is valid
		err = sg.Execute(operation)
		check(err)

		// creating a new device
		go log("recovering account", done)

		resp, err = client.Post("/v1/identities/"+args[0]+"/history", "application/json", operation)
		done <- err

		if err != nil {
			os.Exit(1)
		}

		fmt.Println("")
		if edsk != "" {
			fmt.Println("device private key:    ", dkid+":"+edsk)
			fmt.Println("device public key:     ", edpk)
		}
		if ersk != "" {
			fmt.Println("recovery private key:  ", rkid+":"+ersk)
			fmt.Println("recovery public key:   ", erpk)
		}
	},
}

func init() {
	accountCommand.AddCommand(accountRecoverCommand)
	accountRecoverCommand.Flags().StringVarP(&recoveryKey, "recovery-key", "r", "", "Recovery secet key")
	accountRecoverCommand.Flags().IntVarP(&effectiveFrom, "effective-from", "f", 0, "Unix timestamp denoting when the action takes effect")
	accountRecoverCommand.Flags().StringVarP(&devicePublicKey, "device-public-key", "p", "", "Device public key")
	accountRecoverCommand.Flags().StringVarP(&recoveryPublicKey, "device-recovery-key", "q", "", "Recovery public key")
}

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

	"github.com/joinself/self-go-sdk/pkg/ntp"
	"github.com/joinself/self-go-sdk/pkg/siggraph"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ed25519"
)

var accountRecoverCommand = &cobra.Command{
	Use:   "recover",
	Short: "recover an account with a recovery key",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			check(errors.New("you must specify an app identity and device [appID, deviceID]"))
		}

		if secretKey == "" {
			check(errors.New("You must provide a secret key"))
		}

		dpk, dsk, err := ed25519.GenerateKey(rand.Reader)
		check(err)

		edpk := enc.EncodeToString(dpk)
		edsk := base64.RawStdEncoding.EncodeToString(dsk)

		rpk, rsk, err := ed25519.GenerateKey(rand.Reader)
		check(err)

		erpk := enc.EncodeToString(rpk)
		ersk := base64.RawStdEncoding.EncodeToString(rsk)

		client := rest(args[0], secretKey)

		done := make(chan error)

		// get the identity history
		go log("getting identity history", done)

		resp, err := client.Get("/v1/identities/" + args[0])
		done <- err

		if err != nil {
			os.Exit(1)
		}

		// load the signature graph
		var app Identity

		err = json.Unmarshal(resp, &app)
		check(err)

		sg, err := siggraph.New(app.History)
		check(err)

		// create a new operation
		rkid := strconv.Itoa(len(sg.Keys()) + 1)
		dkid := strconv.Itoa(len(sg.Keys()) + 2)
		ddid := strconv.Itoa(len(sg.Devices()) + 1)

		actions := []siggraph.Action{
			{
				KID:           rkid,
				Type:          siggraph.TypeRecoveryKey,
				Action:        siggraph.ActionKeyAdd,
				EffectiveFrom: ntp.TimeFunc().Unix(),
				Key:           erpk,
			},
			{
				KID:           dkid,
				DID:           ddid,
				Type:          siggraph.TypeDeviceKey,
				Action:        siggraph.ActionKeyAdd,
				EffectiveFrom: ntp.TimeFunc().Unix(),
				Key:           edpk,
			},
		}

		operation := newOperation(sg, actions, secretKey)

		// check the operation is valid
		err = sg.Execute(operation)
		check(err)

		// creating a new device
		go log("revoking device key", done)

		resp, err = client.Post("/v1/identities/"+args[0]+"/history", "application/json", operation)
		done <- err

		if err != nil {
			os.Exit(1)
		}

		fmt.Println("")
		fmt.Println("device private key:    ", dkid+":"+edsk)
		fmt.Println("device public key:     ", edpk)
		fmt.Println("recovery private key:  ", rkid+":"+ersk)
		fmt.Println("recovery public key:   ", erpk)
	},
}

func init() {
	accountCommand.AddCommand(accountRecoverCommand)
	accountRecoverCommand.Flags().StringVarP(&recoveryKey, "--recovery-key", "r", "", "Recovery secet key")
	accountRecoverCommand.Flags().IntVarP(&effectiveFrom, "--effective-from", "f", 0, "Unix timestamp denoting when the action takes effect")
}

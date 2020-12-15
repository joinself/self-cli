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

var deviceCycleCommand = &cobra.Command{
	Use:   "cycle",
	Short: "cycles a devices key",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			check(errors.New("you must specify an app identity and device [appID, deviceID]"))
		}

		if secretKey == "" {
			check(errors.New("You must provide a secret key"))
		}

		var epk string
		var esk string

		if devicePublicKey == "" {
			pk, sk, err := ed25519.GenerateKey(rand.Reader)
			check(err)

			epk = enc.EncodeToString(pk)
			esk = base64.RawStdEncoding.EncodeToString(sk)
		} else {
			epk = devicePublicKey
		}

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
		kid := strconv.Itoa(len(sg.Keys()) + 1)

		okid, err := sg.GetKeyID(args[1])
		check(err)

		actions := []siggraph.Action{
			{
				KID:           okid,
				DID:           args[1],
				Type:          siggraph.TypeDeviceKey,
				Action:        siggraph.ActionKeyRevoke,
				EffectiveFrom: ntp.TimeFunc().Unix(),
			},
			{
				KID:           kid,
				DID:           args[1],
				Type:          siggraph.TypeDeviceKey,
				Action:        siggraph.ActionKeyAdd,
				EffectiveFrom: ntp.TimeFunc().Unix(),
				Key:           epk,
			},
		}

		operation := newOperation(sg, actions, secretKey)

		// check the operation is valid
		err = sg.Execute(operation)
		check(err)

		// revoke old device and create a new device
		go log("revoking old device key and creating new device key", done)

		resp, err = client.Post("/v1/identities/"+args[0]+"/history", "application/json", operation)
		done <- err

		if esk != "" {
			fmt.Println("")
			fmt.Println("device private key:  ", kid+":"+esk)
			fmt.Println("device public key:   ", epk)
		}

		if err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	deviceCommand.AddCommand(deviceCycleCommand)
	deviceCycleCommand.Flags().StringVarP(&secretKey, "secret-key", "s", "", "Device secret key")
	deviceCycleCommand.Flags().StringVarP(&devicePublicKey, "device-public-key", "p", "", "New device public key")
}

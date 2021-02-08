// Copyright 2020 Self Group Ltd. All Rights Reserved.

package cmd

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/joinself/self-go-sdk/pkg/ntp"
	"github.com/joinself/self-go-sdk/pkg/siggraph"
	"github.com/spf13/cobra"
)

var deviceRevokeCommand = &cobra.Command{
	Use:   "revoke",
	Short: "revokes a device permanently",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			check(errors.New("you must specify an app identity and device [appID, deviceID]"))
		}

		if secretKey == "" {
			check(errors.New("You must provide a secret key"))
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

		var ef int64

		if effectiveFrom < 1 {
			ef = ntp.TimeFunc().Unix()
		} else {
			ef = int64(effectiveFrom)
		}

		kid, err := sg.GetKeyID(args[1])
		check(err)

		// create a new operation
		actions := []siggraph.Action{
			{
				KID:           kid,
				Type:          siggraph.TypeDeviceKey,
				Action:        siggraph.ActionKeyRevoke,
				EffectiveFrom: ef,
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
	},
}

func init() {
	deviceCommand.AddCommand(deviceRevokeCommand)
	deviceRevokeCommand.Flags().StringVarP(&secretKey, "secret-key", "s", "", "Device secret key")
	deviceRevokeCommand.Flags().IntVarP(&effectiveFrom, "effective-from", "f", 0, "Unix timestamp denoting when the action takes effect")
}

// Copyright 2020 Self Group Ltd. All Rights Reserved.

package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

var deviceActivateCommand = &cobra.Command{
	Use:   "activate",
	Short: "activates a device",
	Long:  "activates a device and advertises it as available for receiving messages",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			check(errors.New("you must specify an app identity and device [appID, deviceID]"))
		}

		if secretKey == "" {
			check(errors.New("You must provide a secret key"))
		}

		client := rest(args[0], secretKey)

		done := make(chan error)

		go log("advertising new device", done)

		device := []byte(`{"id": "` + args[1] + `", "platform": "sdk", "token": "-"}`)

		_, err := client.Post("/v1/identities/"+args[0]+"/devices", "application/json", device)
		done <- err

		if err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	deviceCommand.AddCommand(deviceActivateCommand)
	deviceActivateCommand.Flags().StringVarP(&secretKey, "secret-key", "s", "", "Device secret key")
}

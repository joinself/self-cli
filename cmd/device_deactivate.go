// Copyright 2020 Self Group Ltd. All Rights Reserved.

package cmd

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
)

var deviceDeactivateCommand = &cobra.Command{
	Use:   "deactivate",
	Short: "deactivates a device",
	Long:  "deactivates a device and marks it as unavailable for receiving messages",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			check(errors.New("you must specify an app identity and device [appID, deviceID]"))
		}

		if secretKey == "" {
			check(errors.New("You must provide a secret key"))
		}

		client := rest(args[0], secretKey)

		done := make(chan error)

		go log("deactivating new device", done)

		_, err := client.Delete("/v1/identities/" + args[0] + "/devices/" + args[1])
		done <- err

		if err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	deviceCommand.AddCommand(deviceDeactivateCommand)
	deviceDeactivateCommand.Flags().StringVarP(&secretKey, "secret-key", "s", "", "Device secret key")
}

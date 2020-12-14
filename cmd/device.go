// Copyright 2020 Self Group Ltd. All Rights Reserved.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deviceCommand = &cobra.Command{
	Use:   "device",
	Short: "the device command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("device called")
	},
}

func init() {
	rootCmd.AddCommand(deviceCommand)
}

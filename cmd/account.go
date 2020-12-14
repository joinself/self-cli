// Copyright 2020 Self Group Ltd. All Rights Reserved.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var accountCommand = &cobra.Command{
	Use:   "account",
	Short: "the account command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("account called")
	},
}

func init() {
	rootCmd.AddCommand(accountCommand)
}

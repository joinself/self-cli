// Copyright 2020 Self Group Ltd. All Rights Reserved.

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"

	"github.com/joinself/self-go-sdk/pkg/siggraph"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var deviceListCommand = &cobra.Command{
	Use:   "list",
	Short: "lists all device",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			check(errors.New("you must specify an app identity [appID]"))
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

		// get the devices
		go log("getting devices", done)

		resp, err = client.Get("/v1/identities/" + args[0] + "/devices")
		done <- err

		if err != nil {
			os.Exit(1)
		}

		var deviceArray []string

		err = json.Unmarshal(resp, &deviceArray)
		check(err)

		devices := make(map[string]struct{})

		for _, d := range deviceArray {
			devices[d] = struct{}{}
		}

		dl := deviceList(sg.Devices())
		sort.Sort(dl)

		lines := make([][]string, len(dl))

		for i, d := range dl {
			_, listed := devices[d]

			_, err := sg.ActiveDevice(d)

			if d == "" {
				continue
			}

			if listed && err == nil { //
				lines[i] = []string{d, "\033[1;32m✓", "✓\033[0m"}
			}

			if !listed && err == nil {
				lines[i] = []string{d, "\033[1;31m✘", "\033[1;32m✓\033[0m"}
			}

			if listed && err != nil {
				lines[i] = []string{d, "\033[1:32m✓", "\033[1;31m✘\033[0m"}
			}

			if !listed && err != nil {
				lines[i] = []string{d, "\033[1;31m✘", "\033[1;31m✘\033[0m"}
			}
		}

		fmt.Println("")

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"DID", "PUBLISHED", "ACTIVE"})
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeaderLine(false)
		table.SetRowLine(false)
		table.SetBorder(false)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.AppendBulk(lines)
		table.Render()
	},
}

func init() {
	deviceCommand.AddCommand(deviceListCommand)
	deviceListCommand.Flags().StringVarP(&secretKey, "secret-key", "s", "", "Device secret key")
}

type deviceList []string

func (d deviceList) Len() int {
	return len(d)
}

func (d deviceList) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}

func (d deviceList) Less(i, j int) bool {
	is, ierr := strconv.Atoi(d[i])
	js, jerr := strconv.Atoi(d[j])

	if ierr == nil && jerr == nil {
		return is < js
	}

	return d[i] < d[j]
}

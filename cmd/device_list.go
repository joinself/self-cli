// Copyright 2020 Self Group Ltd. All Rights Reserved.

package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/joinself/self-go-sdk/pkg/siggraph"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var deviceListCommand = &cobra.Command{
	Use:   "list",
	Short: "lists all devices",
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

		kl := deviceList(sg.Keys())
		sort.Sort(kl)

		lines := make([][]string, len(kl))

		for i, k := range kl {
			if k == "" {
				continue
			}

			did, err := sg.GetDeviceID(k)
			if err != nil {
				if err == siggraph.ErrNotDeviceKey {
					continue
				}
				check(err)
			}

			_, active := devices[did]

			ra, err := sg.RevokedAt(k)
			check(err)

			lines[i] = []string{k, did}

			if active {
				lines[i] = append(lines[i], "\033[1;32m✓\033[0m")
			} else {
				lines[i] = append(lines[i], "\033[1;31m✘\033[0m")
			}

			if ra == 0 {
				lines[i] = append(lines[i], "\033[1;34m-\033[0m")
			} else {
				lines[i] = append(lines[i], fmt.Sprintf("\033[1;31m%s\033[0m", time.Unix(ra, 0).Format(time.RFC3339)))
			}

		}

		fmt.Println("")

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"KID", "DID", "ACTIVE", "REVOKED"})
		table.SetAlignment(tablewriter.ALIGN_CENTER)
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

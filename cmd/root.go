// Copyright 2020 Self Group Ltd. All Rights Reserved.

package cmd

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joinself/self-go-sdk/pkg/ntp"
	"github.com/joinself/self-go-sdk/pkg/pki"
	"github.com/joinself/self-go-sdk/pkg/siggraph"
	"github.com/joinself/self-go-sdk/pkg/transport"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/square/go-jose"
	"github.com/tj/go-spin"
	"golang.org/x/crypto/ed25519"
)

var (
	v                 *viper.Viper
	enc               = base64.RawURLEncoding
	secretKey         string
	recoveryKey       string
	devicePublicKey   string
	recoveryPublicKey string
	effectiveFrom     int
)

// Identity represents an identity
type Identity struct {
	SelfID  string            `json:"self_id"`
	Type    string            `json:"type"`
	History []json.RawMessage `json:"history"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "self-cli",
	Short: "CLI for interacting with the Self network",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	v = viper.New()

	v.SetDefault("self_env", "production")
	v.AutomaticEnv()
}

func rest(selfID, sk string) *transport.Rest {
	cfg := transport.RestConfig{
		APIURL:     apiURL(),
		Client:     &http.Client{},
		SelfID:     selfID,
		KeyID:      parseKeyID(sk),
		PrivateKey: parseSecretKey(sk),
	}

	client, err := transport.NewRest(cfg)
	check(err)

	return client
}

func pk(rest *transport.Rest) *pki.Client {
	client, err := pki.New(pki.Config{Transport: rest})
	check(err)

	return client
}

func parseKeyID(sk string) string {
	return strings.Split(sk, ":")[0]
}

func parseSecretKey(sk string) ed25519.PrivateKey {
	kp := strings.Split(sk, ":")
	if len(kp) < 2 {
		check(errors.New("provided secret key is not valid"))
	}

	seed, err := base64.RawStdEncoding.DecodeString(kp[1])
	check(err)

	return ed25519.NewKeyFromSeed(seed)
}

func apiURL() string {
	if v.GetString("self_env") != "" && v.GetString("self_env") != "production" {
		return "https://api." + v.GetString("self_env") + ".joinself.com"
	}

	return "https://api.joinself.com"
}

func log(message string, done chan error) {
	s := spin.New()

	for {
		select {
		case err := <-done:
			if err != nil {
				fmt.Printf("\r  \033[1;31m%s   \033[0m%s\n", "✘", message)
				fmt.Printf("\nerrored with: \n  %s\n", err.Error())
			} else {
				fmt.Printf("\r  \033[1;32m%s   \033[0m%s\n", "✓", message)
			}
			return
		default:
			fmt.Printf("\r  \033[34m%s   \033[0m%s", s.Next(), message)
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func newOperation(sg *siggraph.SignatureGraph, actions []siggraph.Action, sk string) json.RawMessage {
	kp := strings.Split(sk, ":")

	op := &siggraph.Operation{
		Sequence:  sg.NextSequence(),
		Version:   "1.0.0",
		Previous:  sg.PreviousSignature(),
		Timestamp: ntp.TimeFunc().Unix(),
		Actions:   actions,
	}

	data, err := json.Marshal(op)
	check(err)

	opts := &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]interface{}{
			"kid": kp[0],
		},
	}

	seed, err := base64.RawStdEncoding.DecodeString(kp[1])
	check(err)

	dsk := ed25519.NewKeyFromSeed(seed)

	s, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.EdDSA, Key: dsk}, opts)
	check(err)

	jws, err := s.Sign(data)
	check(err)

	return json.RawMessage(jws.FullSerialize())
}

func check(err error) {
	if err != nil {
		fmt.Printf("\nerrored with:\n  %s\n", err.Error())
		os.Exit(1)
	}
}

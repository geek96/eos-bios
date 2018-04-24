// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	bios "github.com/eoscanada/eos-bios"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var biosConfig *bios.Config

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "eos-bios",
	Short: "A tool to launch EOS.IO Software-based networks",
	Long: `This tools allows orchestration of community launches for the mainnet,
for test networks, for in-house networks bootstrapping as well as local development nodes.`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "./config.yaml", "config file")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Use config file from the flag.
	viper.SetConfigFile(cfgFile)

	viper.SetEnvPrefix("EOS_BIOS")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	var err error
	biosConfig, err = buildConfig()
	if err != nil {
		fmt.Println("Error reading configuration:", err)
		os.Exit(1)
	}
}

func buildConfig() (*bios.Config, error) {
	c := &bios.Config{}
	c.Peer.MyAccount = viper.GetString("peer.my_account")
	c.Peer.APIAddress = viper.GetString("peer.api_address")

	if c.Peer.APIAddress != "" {
		u, err := url.Parse(c.Peer.APIAddress)
		if err != nil {
			return nil, fmt.Errorf("api_address: %s", err)
		}
		c.Peer.APIAddressURL = u
	}
	c.Peer.SecretP2PAddress = viper.GetString("peer.secret_p2p_address")
	c.Peer.BlockSigningPrivateKeyPath = viper.GetString("peer.block_signing_private_key_path")
	if c.Peer.BlockSigningPrivateKeyPath != "" {
		privKey, err := ioutil.ReadFile(c.Peer.BlockSigningPrivateKeyPath)
		if err != nil {
			return nil, err
		}

		wif, err := ecc.NewPrivateKey(strings.TrimSpace(string(privKey)))
		if err != nil {
			return nil, err
		}

		c.Peer.BlockSigningPrivateKey = wif
	}

	c.Hooks = map[string]*bios.HookConfig{}
	for _, hook := range bios.ConfiguredHooks {
		hookURL := viper.GetString(fmt.Sprintf("hooks.%s.url", hook.Key))
		hookExec := viper.GetString(fmt.Sprintf("hooks.%s.exec", hook.Key))
		hookWait := viper.GetBool(fmt.Sprintf("hooks.%s.wait", hook.Key))
		if hookURL != "" || hookExec != "" || hookWait {
			c.Hooks[hook.Key] = &bios.HookConfig{
				Exec: hookExec,
				URL:  hookURL,
				Wait: hookWait,
			}
		}
	}

	// TODO: do `pgp` struct? not used
	// TODO: MyParameters ? not used either..

	return c, nil
}

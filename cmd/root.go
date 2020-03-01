/*

Copyright (C) 2017-2018  Ettore Di Giacinto <mudler@gentoo.org>
                         Daniele Rondina <geaaru@sabayonlinux.org>

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/

package cmd

import (
	"fmt"
	"os"
	"strings"

	//	"reflect"
	common "github.com/MottainaiCI/mottainai-cli/common"
	event "github.com/MottainaiCI/mottainai-slack/cmd/event"

	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"
	"github.com/spf13/cobra"
	viper "github.com/spf13/viper"

	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
)

const (
	cliName = `Mottainai Slack Bridge
Copyright (c) 2017-2018 Mottainai

Command line interface for Mottainai bridges`

	cliExamples = `$> mottainai-bridge -m http://127.0.0.1:8080 -k token run

$> mottainai-bridge -m http://127.0.0.1:8080 -k token run
`
)

func initConfig(config *setting.Config) {
	// Set env variable
	config.Viper.SetEnvPrefix("MOTT")
	config.Viper.BindEnv("config")
	config.Viper.SetDefault("config", "")
	config.Viper.SetDefault("etcd-config", false)

	config.Viper.AutomaticEnv()

	// Create EnvKey Replacer for handle complex structure
	replacer := strings.NewReplacer(".", "__")
	config.Viper.SetEnvKeyReplacer(replacer)

	// Set config file name (without extension)
	config.Viper.SetConfigName(setting.MOTTAINAI_CONFIGNAME)

	config.Viper.SetTypeByDefaultValue(true)
}

func initCommand(rootCmd *cobra.Command, config *setting.Config) {
	var pflags = rootCmd.PersistentFlags()
	v := config.Viper

	pflags.StringP("master", "m", "http://localhost:8080", "MottainaiCI webUI URL")
	pflags.StringP("apikey", "k", "fb4h3bhgv4421355", "Mottainai API key")

	pflags.StringP("profile", "p", "", "Use specific profile for call API.")
	pflags.StringP("slackapi", "s", "", "Slack API key")
	pflags.StringP("slackchannel", "", "", "Slack Channel ID")

	v.BindPFlag("master", rootCmd.PersistentFlags().Lookup("master"))
	v.BindPFlag("apikey", rootCmd.PersistentFlags().Lookup("apikey"))
	v.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	v.BindPFlag("slackapi", rootCmd.PersistentFlags().Lookup("slackapi"))
	v.BindPFlag("slackchannel", rootCmd.PersistentFlags().Lookup("slackchannel"))

	pflags.StringP("config", "c", "/etc/mottainai/mottainai-server.yaml",
		"Mottainai Server configuration file or Etcd path")
	pflags.BoolP("remote-config", "r", false,
		"Enable etcd remote config provider")
	pflags.StringP("etcd-endpoint", "e", "http://127.0.0.1:4001",
		"Etcd Server Address")
	pflags.String("etcd-keyring", "",
		"Etcd Keyring (Ex: /etc/secrets/mykeyring.gpg)")

	v.BindPFlag("config", pflags.Lookup("config"))
	v.BindPFlag("etcd-config", pflags.Lookup("remote-config"))
	v.BindPFlag("etcd-endpoint", pflags.Lookup("etcd-endpoint"))
	v.BindPFlag("etcd-keyring", pflags.Lookup("etcd-keyring"))

	rootCmd.AddCommand(
		event.NewEventCommand(config),
	)
}

const (
	srvName = `Mottainai Bridge
Copyright (c) 2017-2019 Mottainai

Mottainai bridge library`

	srvExamples = `$> mottainai-bridge event run -c mottainai-server.yaml`
)

func Execute() {
	// Create Main Instance Config object
	var config *setting.Config = setting.NewConfig(nil)

	initConfig(config)

	var rootCmd = &cobra.Command{
		Short:        srvName,
		Version:      setting.MOTTAINAI_VERSION,
		Example:      srvExamples,
		Args:         cobra.OnlyValidArgs,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			var pwd string
			var v *viper.Viper = config.Viper
			v.AutomaticEnv()

			if v.GetBool("etcd-config") {
				if v.Get("etcd-keyring") != "" {
					v.AddSecureRemoteProvider("etcd", v.GetString("etcd-endpoint"),
						v.GetString("config"), v.GetString("etcd-keyring"))
				} else {
					v.AddRemoteProvider("etcd", v.GetString("etcd-endpoint"),
						v.GetString("config"))
				}
				v.SetConfigType("yml")
			} else {
				if v.Get("config") == "" {
					// Set config path list
					pwd, err = os.Getwd()
					utils.CheckError(err)
					v.AddConfigPath(pwd)
					v.AddConfigPath(setting.MOTTAINAI_CONFIGPATH)
				} else {
					v.SetConfigFile(v.Get("config").(string))
				}
			}

			// Parse configuration file
			err = config.Unmarshal()
			utils.CheckError(err)
			if v.Get("profiles") != nil && !cmd.Flag("master").Changed {

				// PRE: profiles contains a map
				//      map[
				//        <NAME_PROFILE1>:<PROFILE INTERFACE>
				//        <NAME_PROFILE2>:<PROFILE INTERFACE>
				//     ]

				var conf common.ProfileConf
				var profile *common.Profile
				if err = v.Unmarshal(&conf); err != nil {
					fmt.Println("Ignore config: ", err)
				} else {
					if v.GetString("profile") != "" {
						profile, err = conf.GetProfile(v.GetString("profile"))

						if profile != nil {
							v.Set("master", profile.GetMaster())
							if profile.GetApiKey() != "" && !cmd.Flag("apikey").Changed {
								v.Set("apikey", profile.GetApiKey())
							}
						} else {
							fmt.Printf("No profile with name %s. I use default value.\n", v.GetString("profile"))
						}
					}
				}

			}
		},
	}

	initCommand(rootCmd, config)

	// Start command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

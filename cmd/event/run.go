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

package event

import (
	"fmt"
	"net/url"
	"path"

	service "github.com/MottainaiCI/mottainai-bridge/pkg/service"
	client "github.com/MottainaiCI/mottainai-server/pkg/client"
	setting "github.com/MottainaiCI/mottainai-server/pkg/settings"
	"github.com/slack-go/slack"
	"gopkg.in/yaml.v2"

	cobra "github.com/spf13/cobra"
	viper "github.com/spf13/viper"
)

func newEventRun(config *setting.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "run [OPTIONS]",
		Short: "Run event listener",
		Args:  cobra.OnlyValidArgs,
		// TODO: PreRun check of minimal args if --json is not present
		Run: func(cmd *cobra.Command, args []string) {
			var v *viper.Viper = config.Viper
			slackApi := v.GetString("slackapi")
			slackChannel := v.GetString("slackchannel")
			bridge := service.NewBridge(client.NewTokenClient(v.GetString("master"), v.GetString("apikey"), config))

			bridge.Listen("task.created", func(c service.TaskMap) {
				fmt.Println("[Task][Create]: ", c)
				taskId := c["ID"].(string)
				api := slack.New(slackApi, slack.OptionDebug(true))

				u, err := url.Parse(v.GetString("master"))
				u.Path = path.Join(u.Path, "/tasks/display/"+taskId)

				content, _ := yaml.Marshal(c)
				attachment := slack.Attachment{
					Pretext: u.String(),
					Text:    string(content),
					// Uncomment the following part to send a field too
					/*
						Fields: []slack.AttachmentField{
							slack.AttachmentField{
								Title: "a",
								Value: "no",
							},
						},
					*/
				}

				_, _, err = api.PostMessage(slackChannel, slack.MsgOptionText("Task created", false), slack.MsgOptionAttachments(attachment))
				if err != nil {
					fmt.Printf("%s\n", err)
				}
			})

			bridge.Listen("task.removed", func(c service.TaskMap) {
				taskId := c["ID"].(string)
				api := slack.New(slackApi, slack.OptionDebug(true))

				u, err := url.Parse(v.GetString("master"))
				u.Path = path.Join(u.Path, "/tasks/display/"+taskId)

				content, _ := yaml.Marshal(c)
				attachment := slack.Attachment{
					Pretext: u.String(),
					Text:    string(content),
					// Uncomment the following part to send a field too
					/*
						Fields: []slack.AttachmentField{
							slack.AttachmentField{
								Title: "a",
								Value: "no",
							},
						},
					*/
				}

				_, _, err = api.PostMessage(slackChannel, slack.MsgOptionText("Task removed", false), slack.MsgOptionAttachments(attachment))
				if err != nil {
					fmt.Printf("%s\n", err)
				}
			})

			bridge.Listen("task.update", func(TaskUpdates *service.TaskUpdate) {

				taskId := TaskUpdates.Task["ID"].(string)
				api := slack.New(slackApi, slack.OptionDebug(true))

				if _, ok := TaskUpdates.Diff["last_update_time"]; ok {
					return
				}

				u, err := url.Parse(v.GetString("master"))
				u.Path = path.Join(u.Path, "/tasks/display/"+taskId)

				content, _ := yaml.Marshal(TaskUpdates.Diff)
				attachment := slack.Attachment{
					Pretext: u.String(),
					Text:    string(content),
					// Uncomment the following part to send a field too
					/*
						Fields: []slack.AttachmentField{
							slack.AttachmentField{
								Title: "a",
								Value: "no",
							},
						},
					*/
				}

				_, _, err = api.PostMessage(slackChannel, slack.MsgOptionText("Task updated", false), slack.MsgOptionAttachments(attachment))
				if err != nil {
					fmt.Printf("%s\n", err)
				}
			})
			bridge.Run()
		},
	}

	return cmd
}

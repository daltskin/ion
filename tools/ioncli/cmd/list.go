// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
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

	"github.com/apcera/termtables"
	"github.com/lawrencegripper/ion/tools/ioncli/helpers"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List events output by your local modules",
	Long: `When using "dev new" or "events start" commands events output by your module locally
	are stored and can be viewed here.`,
	Run: func(cmd *cobra.Command, args []string) {

		events, err := helpers.GetEventsFromStore(savedEventsDir)
		if err != nil {
			fmt.Println("Error retreiving events raised by the module:" + err.Error())
			return
		}

		table := termtables.CreateTable()
		table.AddHeaders("Module", "Event Type", "Event ID")

		for _, v := range events {
			table.AddRow(v.ModuleName, v.EventType, v.EventID)
		}

		fmt.Println(table.Render())
	},
}

func init() {
	eventsCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

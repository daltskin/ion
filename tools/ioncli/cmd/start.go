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
	"os"
	"path/filepath"

	"github.com/lawrencegripper/ion/tools/ioncli/helpers"
	"github.com/lawrencegripper/ion/tools/ioncli/types"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a new module using the event output of a previous run. Use 'ioncli dev events list' to get a valid eventid",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		savedEvents, err := helpers.GetEventsFromStore(savedEventsDir)
		if err != nil {
			fmt.Println("Error getting events from saved store:" + err.Error())
			return
		}

		eventMap := map[string]types.SavedEventInfo{}
		for _, i := range savedEvents {
			eventMap[i.EventID] = i
		}

		event, exists := eventMap[eventID]
		if !exists {
			fmt.Println("Event you specified doesn't exist in store..")
			return
		}

		fmt.Println("Set the following environment variables before running more module:")
		fmt.Println("	export SHARED_SECRET=dev")
		fmt.Println("	export SIDECAR_BASE_DIR=" + ionBaseDir)
		fmt.Println("	export SIDECAR_PORT=8080")
		os.RemoveAll(filepath.Join(workingDir, ".dev"))
		//cleanup
		err = os.RemoveAll(ionBaseDir)
		if err != nil {
			fmt.Println("Error cleaning up ion base dir:" + err.Error())
			return
		}

		runner, err := helpers.NewSidecarRunnerFromEvent(SidecarBinaryPath, ionBaseDir, workingDir, moduleName, publishesEvents, event)
		if err != nil {
			fmt.Println("Error creating sidecar runner:" + err.Error())
			return
		}

		err = runner.Start()
		if err != nil {
			fmt.Println("Error starting sidecar runner:" + err.Error())
			return
		}

		fmt.Println("Sidecar is running on 'http://localhost:8080' it requries a header of 'secret:dev'")
		fmt.Println("Run your module code now, we'll wait for you to call 'Done' before proceeding")

		output, err := runner.Wait()
		if err != nil {
			fmt.Println("Error waiting for sidecar to complete:" + err.Error())
			return
		}

		fmt.Println("'Done' Called on sidecar")
		fmt.Println("Sidecar Logs:")
		fmt.Println(output)

		events, err := helpers.GetEventsFromDev(filepath.Join(workingDir, ".dev"))
		if err != nil {
			fmt.Println("Error retreiving events raised by the module:" + err.Error())
			return
		}

		fmt.Println(fmt.Sprintf("Module raised %v events.... saving these", len(events)))

		for _, e := range events {
			err = helpers.SaveEvent(moduleName, ionBaseDir, savedEventsDir, e)
			if err != nil {
				fmt.Println("Error saving event:" + err.Error())
				return
			}
		}
	},
}

var (
	eventID string
)

func init() {
	eventsCmd.AddCommand(startCmd)

	startCmd.Flags().StringVar(&eventID, "eventid", "", "The eventid to run with from the saved events")
	startCmd.MarkFlagRequired("eventid")
	startCmd.Flags().StringVar(&publishesEvents, "publishesevents", "", "A CSV seperated list of events this module can raise. Eg publishesevents=face_detected,eye_detected")
	startCmd.Flags().StringVar(&moduleName, "modulename", "", "The name of the module you will run")
	startCmd.MarkFlagRequired("publishesevents")
	startCmd.MarkFlagRequired("modulename")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

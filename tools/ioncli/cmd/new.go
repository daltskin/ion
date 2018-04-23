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
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Used to start a new local ION pipeline with an input file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		fileinfo, err := os.Stat(inputfolder)
		if err != nil {
			fmt.Println("Input file error:" + err.Error())
			return
		}
		if fileinfo.IsDir() {
			fmt.Println("Expected input to be a directory")
		}

		ionBaseDir := filepath.Join(IonCliDir, ".ionBaseDir")
		fmt.Println("Set the following environment variables before running more module:")
		fmt.Println("	export SHARED_SECRET=dev")
		fmt.Println("	export ION_BASE_DIR=" + ionBaseDir)
		fmt.Println("	export SIDECAR_PORT=8080")
		//cleanup
		err = os.RemoveAll(ionBaseDir)
		if err != nil {
			fmt.Println("Error cleaning up ion base dir:" + err.Error())
			return
		}

		err = os.MkdirAll(ionBaseDir, 0777)
		if err != nil {
			fmt.Println("Error creating ion base dir:" + err.Error())
			return
		}
		inBlobDir := filepath.Join(ionBaseDir, "in", "blob")

		err = helpers.CopyDir(inputfolder, inBlobDir)
		if err != nil {
			fmt.Println("Error copying input folder:" + err.Error())
			return
		}

		runner, err := helpers.NewBlankSidecar(SidecarBinaryPath, ionBaseDir, IonCliDir, moduleName, publishesEvents)
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

		events, err := helpers.GetEventsFromDev(IonCliDir)
		if err != nil {
			fmt.Println("Error retreiving events raised by the module:" + err.Error())
			return
		}

		fmt.Println("Events raised by your module:")
		fmt.Println(events)
	},
}

var (
	inputfolder     string //
	publishesEvents string //
	moduleName      string //
)

func init() {
	devCmd.AddCommand(newCmd)
	newCmd.Flags().StringVar(&inputfolder, "inputfolder", "", "An initial input folder for your module. This will be the 'blob' folder in 'ion/in/blob'")
	newCmd.Flags().StringVar(&publishesEvents, "publishesevents", "", "A CSV seperated list of events this module can raise. Eg publishesevents=face_detected,eye_detected")
	newCmd.Flags().StringVar(&moduleName, "modulename", "", "The name of the module you will run")
	newCmd.MarkFlagRequired("inputfile")
	newCmd.MarkFlagRequired("publishesevents")
	newCmd.MarkFlagRequired("modulename")
}

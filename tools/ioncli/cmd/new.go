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
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Used to start a new local ION pipeline with an input file",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		inputFlag := cmd.Flag("inputfile")
		inputFilePath := inputFlag.Value.String()
		if inputFilePath == "" {
			fmt.Println("inputfile is required.")
			return
		}

		if _, err := os.Stat(inputFilePath); os.IsNotExist(err) {
			fmt.Println(err)
		}

		// Watcher example

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err)
		}
		defer watcher.Close()

		done := make(chan bool)
		go func() {
			for {
				select {
				case event := <-watcher.Events:
					log.Println("event:", event)
					if event.Op&fsnotify.Write == fsnotify.Write {
						log.Println("modified file:", event.Name)
					}
				case err := <-watcher.Errors:
					log.Println("error:", err)
				}
			}
		}()

		err = watcher.Add("/tmp/foo")
		if err != nil {
			log.Fatal(err)
		}

		files, err := ioutil.ReadDir("./")
		if err != nil {
			log.Fatal(err)
		}

		for _, f := range files {
			fmt.Println(f.Name())
		}

		files, err := filepath.Glob("*")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(files)
	},
}

func init() {
	devCmd.AddCommand(newCmd)
	newCmd.Flags().String("inputfile", "", "An initial input file for your module")
}

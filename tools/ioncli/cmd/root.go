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
	"os/user"
	"path"
	"path/filepath"
	"runtime"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "The CLI tool for interacting with ION",
	Long:  `This lets your run and configure a dev environment.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

var (
	IonCliDir         string
	SidecarBinaryPath string
	ionBaseDir        string
	workingDir        string
	savedEventsDir    string
)

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}

	// Configure directories to use
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
	}
	ionCliDir := path.Join(usr.HomeDir, ".ion")

	sidecarBinary := ""
	switch runtime.GOOS {
	case "windows":
		sidecarBinary = "sidecar.exe"
	default:
		sidecarBinary = "sidecar"
	}

	configCmd.PersistentFlags().StringVar(&SidecarBinaryPath, "sidecarbinary", path.Join(ionCliDir, sidecarBinary), "The location of the sidecar binary")
	configCmd.PersistentFlags().StringVar(&IonCliDir, "ionclidir", ionCliDir, "The location used by the ION cli to store data and configuration")

	ionBaseDir = filepath.Join(IonCliDir, "basedir")
	workingDir = filepath.Join(IonCliDir, "sidecarworkingdir")
	savedEventsDir = filepath.Join(IonCliDir, "savedevents")

	err = os.MkdirAll(savedEventsDir, 0777)
	if err != nil {
		fmt.Println("Error creating ion base dir:" + err.Error())
		return
	}

	err = os.MkdirAll(ionBaseDir, 0777)
	if err != nil {
		fmt.Println("Error creating ion base dir:" + err.Error())
		return
	}
	err = os.MkdirAll(workingDir, 0777)
	if err != nil {
		fmt.Println("Error working dir:" + err.Error())
		return
	}
}

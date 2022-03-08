/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com

*/
package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes user's directories",
	Run: func(cmd *cobra.Command, args []string) {

		// Create the requests dir
		err := os.Mkdir("requests", 0755)
		if err != nil {
			fmt.Println("requests directory already exists")
		} else {
			fmt.Println("Created requests directory")
		}
		createBlankFile := func(name string) error {
			// If file doesn't exist make it
			if _, err := os.Stat(name); errors.Is(err, os.ErrNotExist) {
				_, err := os.Create(name)
				return err
			}
			return nil
		}
		if createBlankFile("./requests/reqs.yaml") == nil {
			fmt.Println("requests/reqs.yaml already exists")
		} else {
			fmt.Println("Created requests/reqs.yaml")
		}

		// Create the logs dir
		err = os.Mkdir("logs", 0755)
		if err != nil {
			fmt.Println("logs directory already exists")
		} else {
			fmt.Println("Created logs directory")
		}

		// Create the data dir
		err = os.Mkdir("data", 0755)
		if err != nil {
			fmt.Println("data directory already exists")
		} else {
			fmt.Println("Created data directory")
		}
		if createBlankFile("./data/cred.yaml") == nil {
			fmt.Println("data/cred.yaml already exists")
		} else {
			fmt.Println("Created data/cred.yaml directory")
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

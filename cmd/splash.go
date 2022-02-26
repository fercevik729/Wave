/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com
*/
package cmd

import (
	"Wave/driver"
	"github.com/spf13/cobra"
	"log"
)

// splashCmd represents the wave command
var splashCmd = &cobra.Command{
	Use:   "splash",
	Short: "Concurrently runs HTTP requests from the specified file for i sets",
	Run: func(cmd *cobra.Command, args []string) {
		requests, keychain := driver.New(requestsFile, credentialsFile)
		log.Printf("Your credentials are %s\n", keychain)
		driver.Splash(iterations, requests, verbose, keychain)
	},
}

func init() {
	rootCmd.AddCommand(splashCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// splashCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

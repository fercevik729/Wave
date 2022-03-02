/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com

*/
package cmd

import (
	"fmt"
	"github.com/fercevik729/Wave/driver"
	"github.com/spf13/cobra"
)

// whirlCmd represents the whirl command
var whirlCmd = &cobra.Command{
	Use:   "whirl",
	Short: "Sequentially runs the HTTP requests from the specified file for i cycles",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting whirl...")
		requests, keychain := driver.New(requestsFile, credentialsFile)
		driver.Whirlpool(iterations, requests, verbose, logFile, keychain)
		fmt.Println("Process completed")
	},
}

func init() {
	rootCmd.AddCommand(whirlCmd)
}

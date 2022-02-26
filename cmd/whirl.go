/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com

*/
package cmd

import (
	"Wave/driver"
	"fmt"
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// whirlCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// whirlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

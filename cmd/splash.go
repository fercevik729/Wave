/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com
*/
package cmd

import (
	"Wave/helpers"
	"fmt"

	"github.com/spf13/cobra"
)

var iterations int

// splashCmd represents the wave command
var splashCmd = &cobra.Command{
	Use:   "splash",
	Short: "Concurrently runs HTTP requests from the specified file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("splash called")

		requests, creds := helpers.New(requestsFile, credentialsFile)
		helpers.Splash(iterations, requests, verbose, creds)
	},
}

func init() {
	rootCmd.AddCommand(splashCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// splashCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	splashCmd.Flags().IntVarP(&iterations, "iterations", "i", 10, "describes how many sets of requests to run")
}

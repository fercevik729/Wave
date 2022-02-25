/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// splashCmd represents the wave command
var splashCmd = &cobra.Command{
	Use:   "splash",
	Short: "Concurrently runs HTTP requests from the specified file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("splash called")
		fmt.Println(rootCmd.PersistentFlags().Lookup("requestsFile").Value)
	},
}

func init() {
	rootCmd.AddCommand(splashCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// splashCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// splashCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

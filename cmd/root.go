/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com

*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"

	"github.com/spf13/viper"
)

// TODO: fix viper so it binds cfgFile values to rootCmd flags
var cfgFile string
var verbose bool

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "wave",
	Short: "Wave is an automated RESTful API tester",
	Long: `Wave is a command line application that provides multiple options to automatically test your RESTful API 
from the shell interface. It provides an option to concurrently load test your API as well as an option to cyclically test your API.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.PersistentFlags().Lookup("requestsFile").Value)
		fmt.Println("Welcome to wave")
	},
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

	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "wave.yaml", "config file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "option to enable more detailed output")
	rootCmd.PersistentFlags().StringP("requestsFile", "r", "../requests/http.txt", "file containing the HTTP requests")
	rootCmd.PersistentFlags().StringP("token", "t", "", "api token for request authorization")
	err := viper.BindPFlag("requestsFile", rootCmd.PersistentFlags().Lookup("token"))
	err = viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.SetDefault("requestsFile", "../requests/http.txt")

	if err != nil {
		return
	}

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	}
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

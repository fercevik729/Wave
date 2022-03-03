/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var encrypt bool

// protectCmd represents the protect command
var protectCmd = &cobra.Command{
	Use:   "protect",
	Short: "Encrypts and decrypts the credentials file",
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.PersistentFlags().Lookup("pass").Value.String() == "" {
			fmt.Println("Please specify a passphrase")
		} else if encrypt {
			fmt.Println("Encrypting credentials file...")
			// Call encrypt function
		} else if !encrypt {
			fmt.Println("Decrypting credentials file")
			// Call decrypt function
		}
	},
}

func init() {
	rootCmd.AddCommand(protectCmd)

	// Persistent flags for protectCmd
	protectCmd.PersistentFlags().BoolVarP(&encrypt, "encrypt", "e", false, "Pass it with a passphrase to "+
		"encrypt a credentials file. Leave empty to decrypt with the passphrase")
	protectCmd.PersistentFlags().StringP("pass", "p", "", "Passphrase to encrypt/decrypt "+
		"credentials file with.")
}

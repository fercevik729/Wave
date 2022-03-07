/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com

*/
package cmd

import (
	"fmt"
	"github.com/fercevik729/Wave/driver"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
)

var encrypt bool

// protectCmd represents the protect command
var protectCmd = &cobra.Command{
	Use:   "protect",
	Short: "Encrypts and decrypts the credentials file",
	Run: func(cmd *cobra.Command, args []string) {
		key := cmd.PersistentFlags().Lookup("pass").Value.String()
		keyfile := cmd.PersistentFlags().Lookup("keyfile").Value.String()
		if key == "" && keyfile != "" {
			byteKey, e := ioutil.ReadFile(keyfile)
			if e != nil {
				log.Fatalf("%v Make sure your filepath is correct.\n", e)
			}
			key = string(byteKey)
		}
		if key == "" || len(key) < 16 {
			fmt.Println("Please specify a passphrase that is 16 characters long")
			// Call encrypt function
		} else if encrypt {
			fmt.Println("Encrypting credentials file...")
			err := driver.Encrypt(credentialsFile, key)
			if err != nil {
				log.Fatalf("Couldn't encrypt file, err: %v\n", err)
			}
			fmt.Println("Process complete")
		} else if !encrypt {
			fmt.Println("Decrypting credentials file")
			// Call decrypt function
			err := driver.Decrypt(credentialsFile, key)
			if err != nil {
				log.Fatalf("Couldn't decrypt file, err: %v\n", err)
			}
			fmt.Println("Process complete")
		}
	},
}

func init() {
	rootCmd.AddCommand(protectCmd)

	// Persistent flags for protectCmd
	protectCmd.PersistentFlags().BoolVarP(&encrypt, "encrypt", "e", false, "Pass it with a passphrase to "+
		"encrypt a credentials file. Omit this flag to decrypt with the passphrase")
	protectCmd.PersistentFlags().StringP("pass", "p", "", "Passphrase to encrypt/decrypt "+
		"credentials file with.")
	protectCmd.PersistentFlags().StringP("keyfile", "k", "", "Key file with passphrase")
}

/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com

*/
package driver

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
	"io/ioutil"
	"log"
	"os"
)

type KeyError struct{}

func (k *KeyError) Error() string {
	return "Key length is too short"
}

// Encrypt encrypts the contents of a file using a 16 byte long key
// It overwrites the contents of the file at the filepath
func Encrypt(filepath string, key string) error {

	// Check the length
	if len(key) < 16 {
		return &KeyError{}
	}
	// Read file contents
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)
	if err != nil {
		return err
	}
	weakText, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	// Create new cipher
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return err
	}

	// Encrypt text
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err
	}
	nonce := make([]byte, gcm.NonceSize())

	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	strongText := gcm.Seal(nonce, nonce, weakText, nil)

	// Output the text
	err = ioutil.WriteFile(filepath, strongText, 0777)
	if err != nil {
		return err
	}

	return nil

}

// Decrypt decrypts the contents of a file using a 16 byte long key
// It overwrites the contents of the file at the filepath
func Decrypt(filepath string, key string) error {

	// Check the length
	if len(key) < 16 {
		return &KeyError{}
	}
	// Read file contents
	f, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)
	if err != nil {
		return err
	}
	strongText, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	// Create cipher
	c, err := aes.NewCipher([]byte(key))
	if err != nil {
		return err
	}

	// Decrypt the file contents
	gcm, err := cipher.NewGCM(c)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()
	if len(strongText) < nonceSize {
		return err
	}

	nonce, strongText := strongText[:nonceSize], strongText[nonceSize:]
	weakText, err := gcm.Open(nil, nonce, strongText, nil)
	if err != nil {
		return err
	}
	// Output the text
	err = ioutil.WriteFile(filepath, weakText, 0777)
	if err != nil {
		return err
	}
	return nil
}

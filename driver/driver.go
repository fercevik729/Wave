/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com

*/
package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// TODO: Create yaml requests file

// New creates new Request structs and returns a Keychain struct
func New(reqFile, authFile string) (map[string]Request, KeyChain) {

	// Open yaml file
	f, err := os.Open(reqFile)
	if err != nil {
		log.Fatal(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(f)
	// Get the credentials
	credentials := ReadCredentials(authFile)

	// Unmarshal yaml data
	data, _ := ioutil.ReadAll(f)
	var reqs map[string]Request
	err = yaml.Unmarshal(data, &reqs)
	if err != nil {
		log.Fatal(err)
	}

	// Set request bodies
	for _, request := range reqs {
		if request.DataFile != "" {
			request.Body = *ReadJsonFile(request.DataFile)
		}
	}
	return reqs, credentials

}

// Splash runs the specified requests concurrently with the option to count how many requests had a status code of
func Splash(its int, reqs map[string]Request, verbose bool, dest string, chain KeyChain) int {

	// If a destination log file is specified set it as the output otherwise stick with stdout
	if dest != "" {
		outFile, err := os.OpenFile("./logs/"+dest, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Fatalf("Couldn't open output file %s", dest)
		}
		defer func(outFile *os.File) {
			err := outFile.Close()
			if err != nil {
				log.Fatalf("Couldn't close the file %s", dest)
			}
		}(outFile)

		log.SetOutput(outFile)
	}

	log.Printf("Running %d Request(s) for %d sets:\n", len(reqs), its)
	start := time.Now()
	successes := safeCounter{}
	client := &http.Client{}
	var wg sync.WaitGroup

	// Run the requests for its sets and
	for i := 0; i < its; i++ {
		for _, req := range reqs {
			wg.Add(1)
			req := req
			// Create goroutines for each request
			go func() {
				defer wg.Done()
				if verbose {
					successes.Lock()
					defer successes.Unlock()
				}
				r, err := req.PrepareRequest(chain)
				if err != nil {
					log.Fatalf("Couldn't construct %s\n", req)
				}
				// Set other parameters besides the Request body
				resp, err := client.Do(r)
				if err != nil {
					log.Fatalf("%s timed out\n", req)
				}
				// Log using common log format
				code := resp.StatusCode
				log.Printf("%s %d %d\n", req, code, resp.ContentLength)

				// If the status code is the same as the expected and verbose flag is on increment successes and output the json
				if code == req.SuccessCode && verbose {
					var formattedJSON bytes.Buffer
					body, _ := ioutil.ReadAll(resp.Body)
					err := json.Indent(&formattedJSON, body, "", "    ")
					if err != nil {
						log.Fatalf("Error printing response body for %s\n", req)
					}
					successes.counter++
					log.Printf("Response body: %s\n", formattedJSON.String())
				}
			}()
		}
	}
	wg.Wait()
	log.Printf("Total execution time: %s\n", time.Since(start))
	if verbose {
		log.Printf("%d out of %d successful requests\n", successes.counter, len(reqs)*its)
		return successes.counter
	}

	// If verbose isn't enabled simply return 0
	return 0

}

// Whirlpool runs the specified requests cyclically for a specified number of iterations
func Whirlpool(its int, reqs map[string]Request, verbose bool, dest string, chain KeyChain) int {

	// If a destination log file is specified set it as the output otherwise stick with stdout
	if dest != "" {
		outFile, err := os.OpenFile("./logs/"+dest, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Fatalf("Couldn't open output file %s", dest)
		}
		defer func(outFile *os.File) {
			err := outFile.Close()
			if err != nil {
				log.Fatalf("Couldn't close the file %s", dest)
			}
		}(outFile)

		log.SetOutput(outFile)
	}
	log.Printf("Running %d Request(s) for %d sets:\n", len(reqs), its)
	absStart := time.Now()
	client := &http.Client{}
	successes := 0

	for i := 0; i < its; i++ {
		for _, req := range reqs {
			// Get the prepared requests
			r, err := req.PrepareRequest(chain)
			if err != nil {
				log.Fatalf("Couldn't construct %s\n", req)
			}

			reqStart := time.Now()
			resp, err := client.Do(r)
			if err != nil {
				log.Fatalf("%s timed out\n", req)
			}
			// Log using common log format
			code := resp.StatusCode
			log.Printf("%s %d %d, took %s to process\n", req, code, resp.ContentLength, time.Since(reqStart))

			// Get the API token from the POST response body
			if req.Method == "POST" && req.IsAuth {
				var tokenMap map[string]string
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(body, &tokenMap)
				if err != nil {
					log.Fatal(err)
				}
				chain.Token = "Bearer " + tokenMap["token"]
			}
			// If the status code is 200 and verbose flag is enabled increment successes and output the json
			if code == req.SuccessCode {
				successes++
			}
			if verbose {
				var formattedJSON bytes.Buffer
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Indent(&formattedJSON, body, "", "    ")
				if err != nil {
					log.Fatalf("Error printing response body for %s\n", req)
				}
				log.Printf("Response body: %s\n", formattedJSON.String())
			}
		}
	}

	log.Printf("Total execution time: %s\n", time.Since(absStart))
	if verbose {
		log.Printf("%d out of %d successful requests\n", successes, len(reqs)*its)
	}

	return successes
}

// String outputs Request and YAMLRequest details
func (r Request) String() string {
	return fmt.Sprintf("127.0.0.1 - - [%s] \"%s %s HTTP/1.0\"", time.Now().Format("2/Jan/2006:15:04:05 -0700"),
		r.Method, r.Endpoint)
}

// PrepareRequest returns http.Request structs
func (r *Request) PrepareRequest(key KeyChain) (*http.Request, error) {
	req, err := http.NewRequest(r.Method, r.Base+r.Endpoint, &r.Body)
	if err != nil {
		return &http.Request{}, err
	}
	// Set headers appropriately
	req.Header.Set("Content-Type", r.ContentType)
	if r.IsAuth {
		req.SetBasicAuth(key.User, key.Pass)
	} else if r.RToken {
		req.Header.Set("Authorization", key.Token)
	}

	return req, nil

}

// ReadJsonFile reads in JSON files for Create, Update, and Delete requests
func ReadJsonFile(filepath string) *bytes.Buffer {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Couldn't open json file at %s\n", filepath)
	}
	defer func(jsonFile *os.File) {
		err := jsonFile.Close()
		if err != nil {
			log.Fatalf("Couldn't close json file at %s", filepath)
		}
	}(jsonFile)

	byteValue, _ := ioutil.ReadAll(jsonFile)

	return bytes.NewBuffer(byteValue)
}

// ReadCredentials returns a KeyChain struct from a yaml file
func ReadCredentials(filepath string) KeyChain {

	yamlFile, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer func(yamlFile *os.File) {
		err := yamlFile.Close()
		if err != nil {
			log.Fatalf("Couldn't close yaml file at %s", filepath)
		}
	}(yamlFile)

	data, _ := ioutil.ReadAll(yamlFile)
	keys := KeyChain{}
	err = yaml.Unmarshal(data, &keys)

	if err != nil {
		log.Fatalf("%v", err)
	}

	return keys
}

type safeCounter struct {
	sync.Mutex
	counter int
}

type KeyChain struct {
	User  string `yaml:"user"`
	Pass  string `yaml:"pass"`
	Token string `yaml:"token"`
}
type Request struct {
	Method      string `yaml:"method"`
	Base        string `yaml:"base"`
	Endpoint    string `yaml:"endpoint"`
	SuccessCode int    `yaml:"success-code"`
	DataFile    string `yaml:"data-file"`
	ContentType string `yaml:"content-type"`
	IsAuth      bool   `yaml:"is-auth"`
	RToken      bool   `yaml:"r-token"`
	Body        bytes.Buffer
}

func (c *KeyChain) String() string {
	return fmt.Sprintf("Your username: %s, Your password: %s, Your token: %s", c.User, c.Pass, c.Token)
}

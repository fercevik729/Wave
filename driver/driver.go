/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com

*/
package driver

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// New creates new Request structs
func New(inFile string, authFile string) ([]Request, KeyChain) {

	// Gather all the requests and Request data
	f, err := os.Open(inFile)
	if err != nil {
		log.Fatal(err)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Fatalf("Couldn't close the file %s", inFile)
		}
	}(f)

	// Create slice of Request structs
	var requests []Request
	credentials := ReadCredentials(authFile)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.Split(scanner.Text(), " ")
		// Creates the Request struct
		r := Request{
			reqType:  text[0],
			endpoint: text[1],
		}
		if len(text) == 3 && text[2] != "AUTH" {
			r.body = *readJsonFile(text[2])
		} else if len(text) == 3 && text[2] == "AUTH" {
			r.isAuth = true
		}

		requests = append(requests, r)
	}

	return requests, credentials
}

// Splash runs the specified requests concurrently with the option to count how many requests had a status code of
func Splash(its int, reqs []Request, verbose bool, dest string, chain KeyChain) int {

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

	for i := 0; i < its; i++ {
		wg.Add(1)
		go func() {
			if verbose {
				successes.Lock()
				defer successes.Unlock()
			}
			defer wg.Done()
			for _, req := range reqs {
				r, err := http.NewRequest(req.reqType, req.endpoint, &req.body)
				if err != nil {
					log.Fatalf("Couldn't construct %s\n", req)
				}
				// Set other parameters besides the Request body
				r.Header.Set("content-Type", "application/json")
				r.Header.Set("Authorization", chain.Token)
				// If the method is a POST method that gets an API token set the username and password fields
				resp, err := client.Do(r)
				if err != nil {
					log.Fatalf("%s timed out\n", req)
				}
				code := resp.StatusCode
				log.Printf("Status code %d for %s\n", code, req)

				// If the status code is 200 and verbose flag is enabled increment successes and output the json
				if code >= 200 && code < 300 && verbose {
					var formattedJSON bytes.Buffer
					body, _ := ioutil.ReadAll(resp.Body)
					err := json.Indent(&formattedJSON, body, "", "    ")
					if err != nil {
						log.Fatalf("Error printing response body for %s\n", req)
					}
					successes.counter++
					log.Printf("Response body: %s\n", formattedJSON.String())
				}

			}
		}()
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
func Whirlpool(its int, reqs []Request, verbose bool, dest string, chain KeyChain) int {

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

			r, err := http.NewRequest(req.reqType, req.endpoint, &req.body)
			if err != nil {
				log.Fatalf("Couldn't construct %s\n", req)
			}
			// Set other parameters besides the Request body
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Authorization", chain.Token)
			// If the method is a POST method that gets an API token set the username and password fields
			if req.reqType == "POST" && req.isAuth {
				r.SetBasicAuth(chain.User, chain.Pass)
			}
			reqStart := time.Now()
			resp, err := client.Do(r)
			if err != nil {
				log.Fatalf("%s timed out\n", req)
			}
			code := resp.StatusCode
			log.Printf("Status code %d for %s\n", code, req)
			log.Printf("Request took %s to process\n\n", time.Since(reqStart))

			// Get the API token from the POST response body
			if req.reqType == "POST" && req.isAuth {
				var tokenMap map[string]string
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(body, &tokenMap)
				if err != nil {
					log.Fatal(err)
				}
				chain.Token = "Bearer " + tokenMap["token"]
			}
			// If the status code is 200 and verbose flag is enabled increment successes and output the json
			if code >= 200 && code < 300 {
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

func (r Request) String() string {
	return fmt.Sprintf("Request type: %s, endpoint: %s", r.reqType, r.endpoint)
}

// readJsonFile reads in JSON files for CUD requests
func readJsonFile(filepath string) *bytes.Buffer {
	jsonFile, err := os.Open(filepath)
	if err != nil {
		log.Fatalf("Couldn't open json file at %s\n", filepath)
		return nil
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

type Request struct {
	reqType  string
	endpoint string
	body     bytes.Buffer
	isAuth   bool
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

func (c *KeyChain) String() string {
	return fmt.Sprintf("Your username: %s, Your password: %s, Your token: %s", c.User, c.Pass, c.Token)
}

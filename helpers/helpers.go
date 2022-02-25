package helpers

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
	credentials := readCredentials(authFile)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.Split(scanner.Text(), " ")

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
func Splash(its int, reqs []Request, verbose bool, chain KeyChain) {

	fmt.Printf("Running %d Request(s) for %d sets:\n", len(reqs), its)
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
				r.Header.Set("Authorization", chain.token)
				// If the method is a POST method that gets an API token set the username and password fields
				resp, err := client.Do(r)
				if err != nil {
					log.Fatalf("%s timed out\n", req)
				}
				code := resp.StatusCode
				log.Printf("Status code %d for %s\n", code, req)

				// If the status code is 200 and verbose flag is enabled increment successes and output the json
				if code == 200 && verbose {
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
	}
}

// Whirlpool runs the specified requests cyclically for a specified number of iterations
func Whirlpool(its int, reqs []Request, verbose bool, chain KeyChain) {
	fmt.Printf("Running %d Request(s) for %d sets:\n", len(reqs), its)
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
			r.Header.Set("Authorization", chain.token)
			// If the method is a POST method that gets an API token set the username and password fields
			if req.reqType == "POST" && req.isAuth {
				r.SetBasicAuth(chain.user, chain.pass)
			}
			reqStart := time.Now()
			resp, err := client.Do(r)
			if err != nil {
				log.Fatalf("%s timed out\n", req)
			}
			code := resp.StatusCode
			log.Printf("Status code %d for %s\n", code, req)
			log.Printf("Request took %s to process\n", time.Since(reqStart))

			// Get the API token from the POST response body
			if req.reqType == "POST" && req.isAuth {
				var tokenMap map[string]string
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(body, &tokenMap)
				if err != nil {
					log.Fatal(err)
				}
				chain.token = "Bearer " + tokenMap["token"]
			}
			// If the status code is 200 and verbose flag is enabled increment successes and output the json
			if code == 200 && verbose {
				var formattedJSON bytes.Buffer
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Indent(&formattedJSON, body, "", "    ")
				if err != nil {
					log.Fatalf("Error printing response body for %s\n", req)
				}
				successes++
				log.Printf("Response body: %s\n", formattedJSON.String())
			}
			fmt.Println()
		}
	}

	log.Printf("Total execution time: %s\n", time.Since(absStart))
	if verbose {
		log.Printf("%d out of %d successful requests\n", successes, len(reqs)*its)
	}
}

func (r Request) String() string {
	return fmt.Sprintf("Request type: %s, endpoint: %s", r.reqType, r.endpoint)
}

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

// readCredentials returns a KeyChain struct from a yaml file
func readCredentials(filepath string) KeyChain {

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
	user  string `yaml:"user"`
	pass  string `yaml:"pass"`
	token string `yaml:"token"`
}

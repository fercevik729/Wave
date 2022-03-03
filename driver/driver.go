/*
Copyright © 2022 Furkan Ercevik ercevik.furkan@gmail.com

*/
package driver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/copier"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// safeCounter is used to count the number of successful requests from concurrent calls to RESTful API endpoints
type safeCounter struct {
	sync.Mutex
	counter int
}

// KeyChain is used to store API credentials that can be referred to by requests in the requests YAML file
type KeyChain struct {
	User  string `yaml:"user"`
	Pass  string `yaml:"pass"`
	Token string `yaml:"token"`
}

// Request is a struct that contains many fields from the net/http Request struct but also some more:
// Base: represents base url of API
// Endpoint: represents the endpoint of an API, utilizes {id} notation if an IdRange is specified
// IdRange: specifies the id range of the request
// SuccessCode: expected status code for the request after it has been called and processed by the API
// DataFile: filepath to JSON file containing POST, PATCH, or DELETE data
// ExpectFile: filepath to JSON file containing expected response body
// IsAuth: specifies if the method is an authentication method
// RToken: specifies if the method requires a token
type Request struct {
	Method       string   `yaml:"method"`
	Base         string   `yaml:"base"`
	Endpoint     string   `yaml:"endpoint"`
	IdRange      []string `yaml:"id-range"`
	SuccessCode  int      `yaml:"success-code"`
	DataFile     string   `yaml:"data-file"`
	ExpectFile   string   `yaml:"expect-file"`
	ContentType  string   `yaml:"content-type"`
	IsAuth       bool     `yaml:"is-auth"`
	RToken       bool     `yaml:"r-token"`
	body         bytes.Buffer
	expectedBody []byte
}

// setToken sets the token field to the parameter token
func (c *KeyChain) setToken(token string) {
	c.Token = "Bearer" + token
}

// String outputs Request details
func (r Request) String() string {
	return fmt.Sprintf("[%s] \"%s %s HTTP/1.1\"", time.Now().Format("2/Jan/2006:15:04:05 -0700"),
		r.Method, r.Endpoint)
}

// String outputs KeyChain details
func (c *KeyChain) String() string {
	return fmt.Sprintf("Your username: %s, Your password: %s, Your token: %s", c.User, c.Pass, c.Token)
}

// New creates new Request structs and returns a Keychain struct
func New(reqFile, authFile string) ([]*Request, *KeyChain) {

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
	credentials := readCredentials(authFile)

	// Unmarshal yaml data into a map of Request pointers
	data, _ := ioutil.ReadAll(f)
	var reqs map[string]*Request
	err = yaml.Unmarshal(data, &reqs)
	if err != nil {
		log.Fatalf("Check the fields in your YAML requests file: %e", err)
	}

	// Set request bodies
	for _, request := range reqs {
		request.Method = strings.ToUpper(request.Method)
		if request.DataFile != "" {
			request.body = *readJsonFile(request.DataFile)
		}
		// Set the expected body of the request
		if request.ExpectFile != "" {
			request.expectedBody = jsonToByte(request.ExpectFile)
		}
	}

	// Unpack any requests with id ranges
	finalReqs := make([]*Request, 0)
	for _, request := range reqs {
		// If the request has an id range unpack the request and append it to the final slice
		if request.IdRange != nil {
			newReqs, err := request.unpackRequests()
			if err != nil {
				log.Fatalf("Improper id range bounds: %e\n", err)
			}
			finalReqs = append(finalReqs, newReqs...)
			// Otherwise, append the original request
		} else {
			finalReqs = append(finalReqs, request)
		}
	}
	return finalReqs, credentials

}

// Splash runs the specified requests concurrently with the option to count how many requests had a status code of
func Splash(its int, reqs []*Request, verbose bool, dest string, chain *KeyChain) int {

	// If a destination log file is specified set it as the output otherwise stick with stdout
	var out *os.File
	if dest != "" {
		outFile, err := os.OpenFile("./logs/"+dest, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
		out = outFile
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

	// Start message
	startMessage := fmt.Sprintf("Sending %d Request(s) for %d sets to", len(reqs), its)
	count := 0
	for _, request := range reqs {
		// If the request doesn't send a request to the same base add it to the starting message
		if !strings.Contains(startMessage, request.Base) {
			startMessage += " " + request.Base
			if count != len(reqs)-1 {
				startMessage += ","
			}
		}
		count++
	}
	log.Println(startMessage)

	start := time.Now()
	successes := safeCounter{}
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
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
				r, err := req.prepareRequest(chain)
				if err != nil {
					log.Fatalf("Couldn't construct %s\n", req)
				}
				// Set other parameters besides the Request body
				reqStart := time.Now()
				resp, err := client.Do(r)
				if err != nil {
					log.Fatalf("%s timed out\n", req)
				}
				// Log to output file or stdout
				code := resp.StatusCode
				message := fmt.Sprintf("%s %d %d, %s\n", req, code, resp.ContentLength, time.Since(reqStart))
				if dest != "" {
					_, err := out.WriteString(message)
					if err != nil {
						log.Fatal(err)
					}
				} else {
					fmt.Print(message)
				}

				// If the status code is the same as the expected and verbose flag is on increment successes and output the json
				if code == req.SuccessCode && verbose {
					var formattedJSON bytes.Buffer
					body, _ := ioutil.ReadAll(resp.Body)
					// If the request has an expected file, check if the response json body matches with the expected body
					err = json.Indent(&formattedJSON, body, "", "    ")
					if err != nil {
						log.Fatalf("Error printing response body for %s\n", req)
					}
					log.Printf("Response body: %s\n", formattedJSON.String())
					if req.ExpectFile != "" {
						match := jsonEqual(req.expectedBody, body)
						if match {
							successes.counter++
						} else {
							log.Println("Response JSON body does NOT match expected JSON body")
						}
					} else {
						successes.counter++
					}
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
func Whirlpool(its int, reqs []*Request, verbose bool, dest string, chain *KeyChain) int {

	// If a destination log file is specified set it as the output otherwise stick with stdout
	var out *os.File
	if dest != "" {
		outFile, err := os.OpenFile("./logs/"+dest, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
		out = outFile
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
	startMessage := ""
	count := 0
	for _, request := range reqs {
		// If the request doesn't send a request to the same base add it to the starting message
		if !strings.Contains(startMessage, request.Base) {
			startMessage += " " + request.Base
			if count != len(reqs)-1 {
				startMessage += ","
			}
		}
		count++
	}
	log.Println(startMessage)
	absStart := time.Now()
	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	successes := 0

	for i := 0; i < its; i++ {
		for _, req := range reqs {
			// Get the prepared request
			r, err := req.prepareRequest(chain)
			if err != nil {
				log.Fatalf("Couldn't construct %s\n", req)
			}

			// Get start time and run the request
			reqStart := time.Now()
			resp, e := client.Do(r)
			if e != nil {
				log.Fatalf("%s timed out\n", req)
			}
			// Log using common log format
			code := resp.StatusCode
			message := fmt.Sprintf("%s %d %d, %s\n", req, code, resp.ContentLength, time.Since(reqStart))
			if dest != "" {
				_, err := out.WriteString(message)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				fmt.Print(message)
			}

			// Get the API token from the POST response body
			if req.Method == "POST" && req.IsAuth {
				var tokenMap map[string]string
				body, _ := ioutil.ReadAll(resp.Body)
				err := json.Unmarshal(body, &tokenMap)
				if err != nil {
					log.Fatal(err)
				}
				chain.setToken(tokenMap["token"])
			}
			var formattedJSON bytes.Buffer
			body, _ := ioutil.ReadAll(resp.Body)
			// If the status codes and bodies match increment successes
			if code == req.SuccessCode {
				if req.ExpectFile != "" {
					if jsonEqual(body, req.expectedBody) {
						successes++
					} else {
						log.Println("Response JSON body does NOT match expected JSON body")
					}
				} else {
					successes++
				}

			}
			// If verbose is enabled output json
			if verbose {
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

// prepareRequest returns http.Request structs with authentication or authorization if needed
func (r *Request) prepareRequest(key *KeyChain) (*http.Request, error) {
	req, err := http.NewRequest(r.Method, r.Base+r.Endpoint, &r.body)
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

// unpackRequests returns a slice of *Request structs for a given Request struct with an IdRange
func (r *Request) unpackRequests() ([]*Request, error) {
	finalRequests := make([]*Request, 0)

	// If the ids aren't numbers or IdRange is greater than 2, iterate over them normally
	upper, err := strconv.Atoi(r.IdRange[1])
	lower, e := strconv.Atoi(r.IdRange[0])
	if err != nil || e != nil || len(r.IdRange) > 2 {
		for _, id := range r.IdRange {
			// Create the new endpoint
			newEndpoint := strings.ReplaceAll(r.Endpoint, "{id}", id)
			newReq := &Request{}
			err := copier.Copy(&newReq, r)
			if err != nil {
				return nil, err
			}
			// Set the endpoint and clear the other fields
			newReq.Endpoint = newEndpoint
			newReq.IdRange = nil

			finalRequests = append(finalRequests, newReq)
		}

		return finalRequests, nil
	}
	// Improper bounds
	if !(upper > lower) {
		var err error
		return nil, err
	}
	for i := lower; i <= upper; i++ {
		// Create the new endpoint
		newEndpoint := strings.ReplaceAll(r.Endpoint, "{id}", strconv.Itoa(i))
		newReq := &Request{}
		err := copier.Copy(&newReq, r)
		if err != nil {
			return nil, err
		}
		// Set the endpoint and clear the other fields
		newReq.Endpoint = newEndpoint
		newReq.IdRange = nil

		finalRequests = append(finalRequests, newReq)
	}

	return finalRequests, nil
}

// jsonEqual compares two slices of JSONified bytes, returns true if they match, otherwise false
func jsonEqual(a, b []byte) bool {
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		fmt.Println("here")
		fmt.Println(j)
		return false
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		fmt.Println("there")
		return false
	}
	return reflect.DeepEqual(j2, j)

}

// readJsonFile reads in JSON files for Create, Update, and Delete requests
func readJsonFile(filepath string) *bytes.Buffer {
	byteValue := jsonToByte(filepath)
	return bytes.NewBuffer(byteValue)
}

// jsonToByte converts a JSON file to a slice of bytes
func jsonToByte(filepath string) []byte {
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
	return byteValue
}

// readCredentials returns a KeyChain struct from a yaml file
func readCredentials(filepath string) *KeyChain {

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

	return &keys
}

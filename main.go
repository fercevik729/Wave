package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func main() {

	// Get command line variables
	iters := flag.Int("iters", 10, "Number of iterations for the set of HTTP requests")
	reqs := flag.String("reqs", "./requests/http.txt", "Source for HTTP requests")
	verbose := flag.Bool("v", false, "Option to display response bodies and number of successful"+
		"requests ")
	flag.Parse()

	runWave(*iters, *reqs, *verbose)

	fmt.Println("")
}

// runWave runs all the specified http requests in the inFile
func runWave(iterations int, inFile string, verbose bool) {

	// Gather all the requests and request data
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

	var requests []request
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.Split(scanner.Text(), " ")

		r := request{
			reqType:  text[0],
			endpoint: text[1],
		}
		if len(text) == 3 {
			r.body = *readJsonFile(text[2])
		}

		requests = append(requests, r)
	}

	runRequests(iterations, requests, verbose)

}

/* runRequests runs the specified requests concurrently with the option to count how many requests had a status code of
200, though at a slower speed
*/
func runRequests(its int, reqs []request, verbose bool) {

	fmt.Printf("Running %d request(s) for %d sets:\n", len(reqs), its)
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
				r.Header.Set("contentType", "application/json")
				resp, err := client.Do(r)
				if err != nil {
					log.Fatalf("%s timed out\n", req)
				}
				code := resp.StatusCode
				log.Printf("Status code %d for %s\n", code, req)
				if code == 200 && verbose {
					body, _ := ioutil.ReadAll(resp.Body)
					successes.counter++
					log.Println(string(body))
				}

			}
		}()
	}
	wg.Wait()
	log.Printf("Time to execute: %s\n", time.Since(start))
	if verbose {
		log.Printf("%d out of %d successful requests\n", successes.counter, len(reqs)*its)
	}
}

// TODO: Whirlpool function to test requests cyclically
// TODO: Configure for API authentication

func (r request) String() string {
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

type request struct {
	reqType  string
	endpoint string
	body     bytes.Buffer
}

type safeCounter struct {
	sync.Mutex
	counter int
}

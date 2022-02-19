package main

import (
	"bufio"
	"flag"
	"fmt"
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
	flag.Parse()

	runWave(*iters, *reqs)

	fmt.Println("")
}

// runWave runs all the specified http requests in the inFile
func runWave(iterations int, inFile string) {

	// Gather all the requests and request data
	f, err := os.Open(inFile)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	var requests []request
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := strings.Split(scanner.Text(), " ")

		// TODO: read body data for post and delete requests
		requests = append(requests, request{
			reqType:  text[0],
			endpoint: text[1],
		})

	}

	runRequests(iterations, requests)

}

func runRequests(its int, reqs []request) {

	fmt.Printf("Running %d requests for %d sets:\n", len(reqs), its)
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < its; i++ {

		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, req := range reqs {

				fmt.Printf("Running request: %s\n", req)
				switch req.reqType {
				case "GET":
					resp, err := http.Get(req.endpoint)
					if err != nil {
						log.Printf("Request %s timed out\n", req)
					} else {
						log.Printf("Status code for %s: %d\n", req, resp.StatusCode)
					}

				default:
					log.Printf("Invalid request: %s\n", req)
				}
			}
		}()
	}
	wg.Wait()
	log.Printf("Time to execute: %s\n", time.Since(start))
}
func (r request) String() string {
	return fmt.Sprintf("Request type: %s, Endpoint: %s", r.reqType, r.endpoint)
}

type request struct {
	reqType  string
	endpoint string
}

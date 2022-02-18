package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {

	// Get command line variables
	iters := flag.Int("iters", 10, "Number of iterations for the set of HTTP requests")
	reqs := flag.String("reqs", "./requests/http.txt", "Source for HTTP requests")
	flag.Parse()

	runLoad(*iters, *reqs)

	fmt.Println("")
}

// runLoad runs all the specified http requests in the inFile
func runLoad(iterations int, inFile string) []io.ReadCloser {

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

		// TODO: read body data
		requests = append(requests, request{
			reqType:  text[0],
			endpoint: text[1],
			body:     text[2],
		})

	}

	return nil
}

type request struct {
	reqType  string
	endpoint string
	body     string
}

# GoWave

GoWave is a Golang library for load-testing RESTful APIs.

## Installation

Download the library as a .zip file and execute the following command

```bash
go build main.go
```

## Usage

```bash
# Concurrently load tests the API for 10 sets and outputs response bodies
./main -n=10 -reqs=./requests/http.txt -v=true

# Cyclically tests requests in the text file 
./main -n=10 -reqs=./requests/http.txt -v=true -w=true

# Sets API username and password and tests requests cyclically
./main -n=10 -u=godeveloper123 -p=ilovegophers -reqs=./requests/http.txt -w=true


# Sets API token and performs cyclic tests
./main -n=10 -t=123njak2ams120982n11xa1x098956785dcpofwe -reqs=./requests/http.txt -w=true
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License
TODO
# Wave

Wave is a command line application built using the Cobra CLI framework to load-test RESTful APIs.

## Installation

```bash
# To get the package
go get github.com/fercevik729/Wave

# To install the CLI tool
go install Wave
```

## Usage

```bash
# To concurrently load test the API use the 'splash' command
wave splash 

# To sequentially test the API use the 'whirl' command
wave whirl

# To output results to a log file use the -o flag
wave splash -o "first.log"

# To set the credentials yaml file use the -c flag
wave whirl -c "./data/mycredentials.yaml"

# To set the iterations use the -i flag
wave splash -i 20 # 20 sets of requests

# To enable verbose output use the -v flag
wave whirl -v

# To set the requests file use the -r flag
wave splash -r "./reqs/firstapirequests.txt"
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## Future ideas
* Enable the creation of test-suites in YAML
* Enable encrypting data files

## License
© Furkan T. Ercevik

This repository is licensed with a [GNU GPLv3](LICENSE) license.
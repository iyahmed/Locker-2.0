# UCSC CSE 247B (Spring Quarter 2025): Locker 2.0

## Introduction

### Locker 2.0 (Current)

The new Locker 2.0 is an extension of "Locker" for CSE 247B in the Spring Quarter of 2025 as a Master's Capstone Project for Ismail Ahmed, that is supposed to ensure truly secure communication between a trusted etcd client and an untrusted honest-but-curious etcd server by encrypting all client data and obscuring all client data access patterns from the server. It is an implementation of the "Encrypted Multi-map that Hides Query, Access, and Volume Patterns" research paper by Alexandra Boldyreva of the Georgia Institute of Technology and Tianxin Tang of the Eindhoven University of Technology. It uses a generic oblivious memory (ORAM) data structure and a generic oblivious dictionary (OMAP) data structure, which will be PathORAM and vORAM+HIRB, as described in the "Path ORAM: An Extremely Simple Oblivious RAM Protocol" and the "Practical Oblivious Map Data Structure with Secure Deletion and History Independence" research papers. The ORAM data structure implementation is a modified version of `epsolute/path-oram` that does not use Redis or AeroSpike. Then, the ORAM will be used as a component of the OMAP that itself will be used as a secure memory access black-box to enable the implementation of an secure and encrypted mini-map (EMM) network protocol that secures all communication between a trusted client and a honest-but-curious server. Locker 2.0 can work for both a plaintext default etcd server (although that would defeat the purpose of Locker 2.0) as well as an encrypted custom etcd server using third-party libraries and software. Locker 2.0 requires a modern Linux server with Bash, Go, and common GNU Linux tools. It is designed to be used cases where there needs to be extremely secret client-server communication and the cost of performance. There is a caveat of all single-value responses being returned as strings but all multiple-value responses being lists and another caveat of read requests not using the "val" field. There is extensive and rigorous testing and benchmarking available. The goal of Locker 2.0 is to enable free and open-source (FOSS) drop-in secure communication for services using etcd, with support for any type of etcd client-server communication (local, Docker, or Internet).

### Original Locker (Outdated)

The original Locker was an implementation, written by Ismail Ahmed and Sallar Farokhi in Golang and Bash for CSE 239A in the Winter Quarter of 2025 as a Final Project, of the Waffle oblivious computation (OC) research paper for the etcd data storage that obscured all data access patterns between a trusted client and an untrusted original server by routing all communications through a trusted third-party proxy. It was adjustable and easily modifiable, with the number of users, requests, and constants being easily updatable using command-line arguments and its API. By default, it did not contain any encryption or authentication for the etcd server itself (as it uses the default etcd server), but that could be easily added by writing a custom etcd config file or by using an external etcd encryption library. It was designed to be used in a Linux server with Bash and was only tested on an Ubuntu 20.04 virtual machine. Its use cases involved any client-server communication that needs to be secret and secure, with the necessary caveats (including the fact that, in a response, a single value is a string but multiple objects are lists). Future aspirations are to extend Locker to Kubernetes or even publish Locker as a Helm module.

## Usage

### Install GVM for Golang (version go1.22.1)

```bash
gvm install go1.4 -B
gvm use go1.4
export GOROOT_BOOTSTRAP=$GOROOT
gvm install go1.17.13
gvm use go1.17.13
export GOROOT_BOOTSTRAP=$GOROOT
gvm install go1.22.1
gvm use go1.22.1
export GOROOT_BOOTSTRAP=$GOROOT
```

### Install Homebrew for etcd (version 3.5.18)

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
brew update
brew install etcd
```

### Installing the Golang Modules (listed in `go.mod`)

```bash
go mod tidy
```

### Installing the Locker 2.0 Dependencies (PathORAM)

```bash
cd libs
chmod +x build.sh
./build.sh
```

### Starting the server and proxy

First Terminal Session:
NOTE: Must be using Version 3.5 (latest) of etcd or a version that supports the v3 API.

```bash
etcd
```

Second Terminal Session (Secure, Default):

```bash
export LD_LIBRARY_PATH=$(pwd)/libs/oram
gvm use go1.22.1 
go run secure.go
```

Second Terminal Session (Plaintext, Stored in `mains/`):

```bash
gvm use go1.22.1 
go run plaintext.go
```

### Running the benchmarks for Locker 2.0

Third Terminal Session (the benchmark's data log file is in `execs/benchmark_data.txt`):

First, set up the shell scripts:

```bash
cd execs/
sudo chmod +x init.sh
sudo chmod +x benchmark.sh
```

Then, we initialize the database with the chosen key text file and the number of values to input (defaults to the entire file):

```bash
./init.sh <small_keys.txt | medium_keys.txt | large_keys.txt> <MAX_VALUES>
```

Lastly, we run the benchmark, which is ran with either `secure.go` or `plaintext.go` (defaults to 10 users):

```bash
./benchmark.sh  -n <NUM_REQUESTS> -b <MAX_BATCH_SIZE> -v <MAX_VALUE_SIZE> -r <READ_PERCENTAGE>
```

To compare the performance with and without security, swap the `secure.go` and `plaintext.go` files in the `Locker-2.0/` main folder.

### GET/PUT Request Formats

This is the format that the curl requests must adhere to in order to use the proxy:

```json
data: [
    {"rid", "op", "key", "val"}
]
```

```bash
curl -s -X POST http://localhost:5000   -H "Content-Type: application/json"   -d '[{"rid": "1", "op": "read", "key": "foo"}]'
curl -s -X POST http://localhost:5000   -H "Content-Type: application/json"   -d '[{"rid": "1", "op": "read", "key": "foo"},{"rid": "2", "op": "read", "key": "bar"}]'
curl -s -X POST http://localhost:5000   -H "Content-Type: application/json"   -d '[{"rid": "1", "op": "write", "key": "foo", "val": "bar"}]'
```

## License

Locker © 2025 by Sallar Farokhi, Ismail Ahmed is licensed under Creative Commons Attribution 4.0 International

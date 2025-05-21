# UCSC CSE 247B (Spring Quarter 2025): Locker 2.0 - An Oblivious ectd Communication Client

## Introduction to Locker 2.0

### Locker 2.0 (Current)

The new Locker 2.0 Oblivious etcd Communication Client (an extension of "Original Locker") is written for CSE 247B in the Spring Quarter of 2025 as a Master's Capstone Project for Ismail Ahmed. Locker 2.0 ensures truly secure communication between a trusted etcd client and an untrusted honest-but-curious network connection and etcd server by encrypting all client data and obscuring all client data access patterns from any non-client eavesdropping adversaries. It is an implementation of the "Encrypted Multi-map that Hides Query, Access, and Volume Patterns" research paper by Alexandra Boldyreva of the Georgia Institute of Technology and Tianxin Tang of the Eindhoven University of Technology, which was presented at the SCN 2024 conference in Amalfi (SA), Italy. It uses a generic oblivious memory (ORAM) data structure and a generic oblivious dictionary/map (OMAP) data structure to form the encrypted multi-map (EMM) oblivious data structure (OBS).

We have chosen PathORAM to be the ORAM and vORAM+HIRB to be the OMAP data structure in our implementation of EMM, as described in the "Path ORAM: An Extremely Simple Oblivious RAM Protocol" and the "Practical Oblivious Map Data Structure with Secure Deletion and History Independence" research papers. The PathORAM data structure is implemented in an C++ library that is statically linked and compiled, based on the `obliviousram/PathORAM` GitHub repository. The major modification made to the `obliviousram/PathORAM` GitHub repository is the block size being 64 bytes, as defined in `Block.h`, to allow for messages up to 64 bytes to be stored in a single block that is large enough to contain short messages while still being small enough to preserve obliviousness. The PathORAM data structure will be used as a component of the vORAM+HIRB data structure, which is implemented in an altered Python library that is inteperted, based on `dsroche/obliv`. The only major modification made to the `dsroche/obliv` GitHub repository is changing the block size to be 64 bytes to be in tune with the PathORAM data structure's implementation. The vORAM+HIRB data structure itself is used as a black-box oblivious key-value data store to enable the from-scratch implementation of an encrypted mini-map (EMM) data structure that secures all communication between a trusted client and a honest-but-curious network connection and server. As to my knowledge, there is no currently active free and open source (FOSS) implementaiton of the encrypted multi-map (EMM) data structure, so we implemented our EMM data structure from scratch and have implemented it in such as way as to fit in well with vORAM+HIRB as its OMAP data structure and PathORAM as its ORAM data structure.

Locker 2.0 can work for both a plaintext default etcd server as well as an encrypted custom etcd server using third-party libraries and software. Locker 2.0 requires a modern Linux environment with Bash support, Go support, common GNU Linux tools installed as well as specialist librarires as per the instructions of `obliviousram/PathORAM` and `dsroche/obliv` on GitHub, and to be ran on a 64-bit computer. There is extensive and rigorous testing and benchmarking scripts available, which are stored in the `execs/` directiory. It is designed to be used in cases where there needs to be extremely secret client-server communication where users do not mind paying the inevitable cost of performance. The goal of Locker 2.0 is to enable free and open-source (FOSS) drop-in secure communication for services using etcd, with support for any type of etcd client-server communication (local, Docker, or Internet), and we have achieved that goal.

There are some caveats with Locker 2.0, however, which should cause caution with API developers for Locker 2.0. All single-value JSON responses are returned as strings but all multiple-value JSON responses are returned as lists. Read JSON requests do not use the `val` field, so the `val` field could be safely ignored by API developers. Each individual message can be to 64 bytes long, although this is adjustable inside the source code of `deps/PathORAM/include/Blocks.h`, `deps/obliv/obliv_server.py`, and `libs/emm/EMM_server.go`. By default, the URLs used inside Locker 2.0 can only be modified inside the source code of `secure.go`, `mains/proxy.go`, `mains/plaintext.go`, `deps/obliv/obliv_server.py`, `libs/hirb/hirb_client.go`, and `libs/hirb/hirb_server.go`. Finally, `secure.go`'s default external HTTP address is `127.0.0.1:5000/etcd` compared to `proxy.go` and `plaintext.go`'s default external HTTP address of `127.0.0.1:5000`.

### Original Locker (Outdated)

The original Locker was an implementation, written by Ismail Ahmed and Sallar Farokhi in Golang and Bash for CSE 239A in the Winter Quarter of 2025 as a Final Project, of the Waffle oblivious computation (OC) research paper for the etcd data storage that obscured all data access patterns between a trusted client and an untrusted original server by routing all communications through a trusted third-party proxy. It was adjustable and easily modifiable, with the number of users, requests, and constants being easily updatable using command-line arguments and its API. By default, it did not contain any encryption or authentication for the etcd server itself (as it uses the default etcd server), but that could be easily added by writing a custom etcd config file or by using an external etcd encryption library. It was designed to be used in a Linux server with Bash and was only tested on an Ubuntu 20.04 virtual machine. Its use cases involved any client-server communication that needs to be secret and secure, with the necessary caveats (including the fact that, in a response, a single value is a string but multiple objects are lists). Future aspirations are to extend Locker to Kubernetes or even publish Locker as a Helm module.

## Building & Installing Locker 2.0

### Installing GVM for Golang (Version go1.22.1)

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

### Installing Homebrew for etcd (Version 3.5.18)

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
brew update
brew install etcd
```

### Installing the Golang Modules for Locker.2.0 (All modules are listed in `go.mod`)

```bash
go mod tidy
```

### Building the PathORAM Dependency for Locker 2.0

```bash
export LD_LIBRARY_PATH=$(pwd)/libs/oram
gvm use go1.22.1
cd libs/oram
chmod +x build.sh
bash build.sh
# Testing: Uncomment out the testORAM() function in `secure.go` and run `secure.go`
```

### Building the vORAM+HIRB Dependency for Locker 2.0

```bash
cd deps/obliv
make clean
make
pip install --force-reinstall --no-binary :all: pycryptodome paramiko
ln -s ~/.local/lib/python3.13/site-packages/Crypto venv/lib/python3.12/site-packages/
# Testing: Run pytest -rs after building or uncomment out the testORAM() function in `secure.go` and run `secure.go`
```

## Using Locker 2.0

### First Usage Terminal

```bash
etcd
```

### Second Usage Terminal (Only needed for `secure.go`)

```bash
cd deps/obliv
python obliv_server.py
```

### Third Usage Terminal (Default high security etcd communication client)

```bash
gvm use go1.22.1
go run secure.go
```

### Third Usage Terminal Session (Optional weak security version of `proxy.go` and some modification is required to work with testing/benchmarking and external API calls written for `secure.go`)

```bash
gvm use go1.22.1 
cd ../mains
mv proxy.go ../
go run proxy.go
```

### Third Usage Terminal Session (Optional plaintext version of `secure.go` and some modification is required to work with testing/benchmarking and external API calls written for `secure.go`)

```bash
gvm use go1.22.1 
cd ../mains
mv plaintext.go ../
go run plaintext.go
```

## Fourth Usage Terminal (This is only needed for `curl/wegt` requests from the terminal or to run another program that uses Locker 2.0)

```bash
# Call any API that is designed to work with either `secure.go`, `proxy.go`, or `plaintext.go`
```

## Benchmarking/Testing Locker 2.0 Locally for `secure.go` (The benchmark's data log file is in `execs/benchmark_data.txt` and it has its own defaults, as seen in `benchmark.sh`)

### First Testing Terminal

```bash
etcd
```

### Second Testing Terminal

```bash
cd deps/obliv
python obliv_server.py
```

### Third Testing Terminal

```bash
gvm use go1.22.1
go run secure.go
```

## Fourth Testing Terminal

```bash
cd execs
chmod +x init.sh
chmod +x benchmarks.sh
bash init.sh <small_keys.txt | medium_keys.txt | large_keys.txt> <MAX_VALUES>
bash benchmark.sh --num-requests NUM --batch-size SIZE --val-size MAX --warmup-batches NUM --read-percentage PCT --key-file FILE --help
```

To compare the performance of `secure.go`, `proxy.go`, and `plaintext.go`, make sure that all three folders are in `Locker-2.0/`'s main folder, modify `benchmarks.sh` as needed, and follow the instructions in the previous Section (except for the Fourth Terminal instruction in this Section).

## API Calls in Locker 2.0

### JSON Data Body Format

```json
data: [
    {"rid", "op", "key", "val"}
]
```

### `curl` API Request Format

```bash
curl -s -X POST http://localhost:5000/etcd   -H "Content-Type: application/json"   -d '[{"rid": "1", "op": "read", "key": "foo"}]'

curl -s -X POST http://localhost:5000/etcd   -H "Content-Type: application/json"   -d '[{"rid": "1", "op": "read", "key": "foo"},{"rid": "2", "op": "read", "key": "bar"}]'

curl -s -X POST http://localhost:5000/etcd   -H "Content-Type: application/json"   -d '[{"rid": "1", "op": "write", "key": "foo", "val": "bar"}]'

curl -s -X POST http://localhost:5000/etcd   -H "Content-Type: application/json"   -d '[{"rid": "1", "op": "delete", "key": "foo"}]'
```

### `wget` API Request Format

```bash
wget --quiet \
     --method=POST \
     --header="Content-Type: application/json" \
     --body-data='[{"rid": "1", "op": "read", "key": "foo"}]' \
     http://localhost:5000/etcd -O -
w
get --quiet \
     --method=POST \
     --header="Content-Type: application/json" \
     --body-data='[{"rid": "1", "op": "read", "key": "foo"},{"rid": "2", "op": "read", "key": "bar"}]' \
     http://localhost:5000/etcd -O -

wget --quiet \
     --method=POST \
     --header="Content-Type: application/json" \
     --body-data='[{"rid": "1", "op": "write", "key": "foo", "val": "bar"}]' \
     http://localhost:5000/etcd -O -

wget --quiet \
     --method=POST \
     --header="Content-Type: application/json" \
     --body-data='[{"rid": "1", "op": "delete", "key": "foo"}]' \
     http://localhost:5000/etcd -O -
```

## Licensing of Locker 2.0

Locker 2.0 © 2025 by Ismail Ahmed and Sallar Farokhi is licensed under Creative Commons Attribution 4.0 International

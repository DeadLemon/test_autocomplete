# Autocomplete Test Assignment

This repository contains a simple TCP autocomplete service implemented in Go.

## Building

Use the Go toolchain to build the server and the client:

```bash
go build -o server ./cmd/server
go build -o client ./cmd/client
```

## Running the Server

Generate a `word_freq.txt` file with words and frequencies (one per line,
`<word> <freq>`). A helper script `mock.py` is provided to create a random file.

```bash
python mock.py          # produces word_freq.txt in the repo root
./server --filename word_freq.txt --port 10000
```

## Running the Client

In a separate terminal:

```bash
./client --host 127.0.0.1 --port 10000
```

Enter commands of the form `get <prefix>` to receive suggestions.

## Testing

The service has unit tests for command parsing, LRU cache and suggestion logic.
Run all tests with:

```bash
go test ./...
```

`go vet` can also be run for static analysis:

```bash
go vet ./...
```


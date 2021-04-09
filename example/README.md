# Plugin example

## Setup

Build and create a testing plugin to be used by the carousel binary.

```bash
go build -buildmode=plugin -o hostValidator.so main.go
```

## Run with Plugin

```bash
go run ./cmd rollout -p example/hostValidator.so 2 0.10.0
```
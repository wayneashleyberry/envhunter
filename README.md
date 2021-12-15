[![Go](https://github.com/wayneashleyberry/envhunter/actions/workflows/go.yml/badge.svg)](https://github.com/wayneashleyberry/envhunter/actions/workflows/go.yml)

> `envhunter` is a static analysis tool for go code that hunts down references to environment variables.

### Installation

```sh
go install github.com/wayneashleyberry/envhunter@latest
```

### Usage

```sh
envhunter ./...
```

### Supported Functions

- [`os.Getenv`](https://pkg.go.dev/os#Getenv)
- [`envconfig.Process`](https://pkg.go.dev/github.com/kelseyhightower/envconfig#Process)
- [`envconfig.MustProcess`](https://pkg.go.dev/github.com/kelseyhightower/envconfig#MustProcess)

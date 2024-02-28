# Fillmore Labs Promise

[![Go Reference](https://pkg.go.dev/badge/fillmore-labs.com/promise.svg)](https://pkg.go.dev/fillmore-labs.com/promise)
[![Build Status](https://badge.buildkite.com/d953ce7905d76167a632cd4d4d0038c13f44abaaa11958795f.svg)](https://buildkite.com/fillmore-labs/promise)
[![GitHub Workflow](https://github.com/fillmore-labs/promise/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/fillmore-labs/promise/actions/workflows/test.yml)
[![Test Coverage](https://codecov.io/gh/fillmore-labs/promise/graph/badge.svg?token=961WZJJOCP)](https://codecov.io/gh/fillmore-labs/promise)
[![Maintainability](https://api.codeclimate.com/v1/badges/12a77c18122e2d1e1f6b/maintainability)](https://codeclimate.com/github/fillmore-labs/promise/maintainability)
[![Go Report Card](https://goreportcard.com/badge/fillmore-labs.com/promise)](https://goreportcard.com/report/fillmore-labs.com/promise)
[![License](https://img.shields.io/github/license/fillmore-labs/promise)](https://www.apache.org/licenses/LICENSE-2.0)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Ffillmore-labs%2Fpromise.svg?type=shield&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Ffillmore-labs%2Fpromise?ref=badge_shield&issueType=license)

The `promise` package provides interfaces and utilities for writing asynchronous code in Go.

## Motivation

Promises and futures are constructs used for asynchronous and concurrent programming, allowing developers to work with
values that may not be immediately available and can be evaluated in a different execution context.

Go is known for its built-in concurrency features like goroutines and channels.
The select statement further allows for efficient multiplexing and synchronization of multiple channels, thereby
enabling developers to coordinate and orchestrate asynchronous operations effectively.
Additionally, the context package offers a standardized way to manage cancellation, deadlines, and timeouts within
concurrent and asynchronous code.

On the other hand, Go's error handling mechanism, based on explicit error values returned from functions, provides a
clear and concise way to handle errors.

The purpose of this package is to provide a library which simplifies the integration of concurrent
code while providing a cohesive strategy for handling asynchronous errors.
By adhering to Go's standard conventions for asynchronous and concurrent code, as well as error propagation, this
package aims to enhance developer productivity and code reliability in scenarios requiring asynchronous operations.

## Usage

Assuming you have a synchronous function `func getMyIP(ctx context.Context) (string, error)` returning your external IP
address (see [GetMyIP](#getmyip) for an example).

Now you can do

```go
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	query := func() (string, error) {
		return getMyIP(ctx)
	}
	future := promise.NewAsync(query)
```

and elsewhere in your program, even in a different goroutine

```go
	if ip, err := future.Await(ctx); err == nil {
		slog.Info("Found IP", "ip", ip)
	} else {
		slog.Error("Failed to fetch IP", "error", err)
	}
```

decoupling query construction from result processing.

### GetMyIP

Sample code to retrieve your IP address:

```go
const serverURL = "https://httpbin.org/ip"

func getMyIP(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, serverURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	ipResponse := struct {
		Origin string `json:"origin"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&ipResponse)

	return ipResponse.Origin, err
}
```

## Links

- [Futures and Promises](https://en.wikipedia.org/wiki/Futures_and_promises) in the English Wikipedia

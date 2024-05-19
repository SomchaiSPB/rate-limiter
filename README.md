# Rate Limiter

This is a simple rate limiter implemented in Go using the standard library. It limits by default:
- 5 messages per second from user
- 10,000 requests per minute per IP address
- 3 failed transactions per user per day

For user limits, please include "X-User-ID" into a request header.

To determine request type: ```message``` or ```transaction```, please include "X-Request-Type" header

Requests with empty or other "X-Request-Type" will be treated as http.StatusOK (200) as long as IP address is allowed.

## Getting started
```shell
go get github.com/SomchaiSPB/rate-limiter
```

## Configuration

The rate limits are defined with default config in the code. You can change these values using config builder:
```go
	conf := NewConfig().
		WithMaxMessages(10).
		WithMaxRequests(10).
		WithMaxFailedTransactions(1)

	rm := NewRateLimiterWithConfig(conf)
	// Rest of the code
```

## Usage

To use the rate limiter, wrap your HTTP handler with the `RateLimiterMiddleware` middleware.

```go
package main

import (
"fmt"
ratelimiter "github.com/SomchaiSPB/rate-limiter"
"log"
"net/http"
)

func main() {
	mux := http.NewServeMux()

	rm := ratelimiter.NewRateLimiter()

	mux.Handle("/", rm.RateLimiterMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Request allowed")
	})))

	log.Fatal(http.ListenAndServe(":8080", mux))
}
```

## Tests

You can run tests using standard go command:
```shell
go test .
```
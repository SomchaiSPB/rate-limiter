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

	port := ":8080"

	log.Println("server started at: " + port)
	log.Fatal(http.ListenAndServe(port, mux))
}

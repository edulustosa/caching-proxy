package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/edulustosa/caching-proxy/memory"
)

func main() {
	port := flag.Int("port", 3000, "Port to listen on")
	origin := flag.String("origin", "", "Origin to proxy to")
	redisURL := flag.String("redis-url", "", "URL to connect to Redis (optional)")

	flag.Parse()

	originURL, err := url.Parse(*origin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "origin must be a valid url")
		os.Exit(1)
	}

	var proxy http.HandlerFunc
	if *redisURL != "" {
		// TODO
	} else {
		cache := memory.NewCache()
		proxy = handleRequest(originURL, cache)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: http.HandlerFunc(proxy),
	}

	log.Printf("Starting proxy server on port %d", *port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

var customTransport = http.DefaultTransport

type cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Clear(ctx context.Context) error
}

func handleRequest(origin *url.URL, cache cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetURL := origin.String() + r.URL.RequestURI()

		data, err := cache.Get(r.Context(), targetURL)
		if err != nil {
			proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
			if err != nil {
				http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
				return
			}

			for name, value := range r.Header {
				for _, v := range value {
					proxyReq.Header.Add(name, v)
				}
			}

			resp, err := customTransport.RoundTrip(proxyReq)
			if err != nil {
				http.Error(w, "Error sending proxy request", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				http.Error(w, "Error reading proxy response", http.StatusInternalServerError)
				return
			}

			if err := cache.Set(r.Context(), targetURL, body); err != nil {
				log.Printf("Error caching response: %v", err)
			}

			for name, values := range resp.Header {
				for _, value := range values {
					w.Header().Add(name, value)
				}
			}

			w.Header().Set("X-Cache", "MISS")
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, bytes.NewReader(body))
		}

		w.Header().Set("X-Cache", "HIT")
		io.Copy(w, bytes.NewReader(data))
	}
}

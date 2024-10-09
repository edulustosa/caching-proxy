package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type ProxyConfig struct {
	Port     int
	Origin   string
	RedisURL *string
}

func main() {
	port := flag.Int("port", 3000, "Port to listen on")
	origin := flag.String("origin", "", "Origin to proxy to")
	redisURL := flag.String("redis-url", "", "URL to connect to Redis (optional)")

	flag.Parse()

	var proxy ProxyConfig
	proxy.Port = *port

	if *origin == "" {
		fmt.Fprintln(os.Stderr, "origin is required")
		os.Exit(1)
	}

	_, err := url.Parse(*origin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "origin must be a valid url")
		os.Exit(1)
	}

	proxy.Origin = *origin

	if *redisURL != "" {
		_, err := url.Parse(*redisURL)
		if err != nil {
			fmt.Fprintln(os.Stderr, "redis-url must be a valid url")
			os.Exit(1)
		}
		proxy.RedisURL = redisURL
	}

	fmt.Println(proxy)
}

var customTransport = http.DefaultTransport

type cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Clear(ctx context.Context) error
}

func handleRequest(targetURL *url.URL, cache cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
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

		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	}
}

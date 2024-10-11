package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/edulustosa/caching-proxy/cache"
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
		memoryCache := cache.NewMemoryCache()
		proxy = handleRequest(originURL, memoryCache)
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

func handleRequest(origin *url.URL, cacheDB cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetURL := origin.String() + r.URL.RequestURI()

		data, err := cacheDB.Get(r.Context(), targetURL)
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

			cacheResp := cache.OriginResponse{
				StatusCode: resp.StatusCode,
				Headers:    resp.Header,
				Body:       string(body),
			}

			if err := cacheDB.Set(r.Context(), targetURL, cacheResp); err != nil {
				log.Printf("Error caching response: %v", err)
			}

			w.Header().Set("X-Cache", "MISS")
			send(w, cacheResp)
		}

		w.Header().Set("X-Cache", "HIT")
		send(w, data)
	}
}

func send(w http.ResponseWriter, resp cache.OriginResponse) {
	if w.Header().Get("X-Cache") == "" {
		for name, values := range resp.Headers {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		w.WriteHeader(resp.StatusCode)
	}

	io.Copy(w, strings.NewReader(resp.Body))
}

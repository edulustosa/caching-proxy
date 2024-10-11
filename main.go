package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/edulustosa/caching-proxy/cache"
	"github.com/edulustosa/caching-proxy/proxy"
)

func main() {
	port := flag.Int("port", 3000, "Port to listen on")
	origin := flag.String("origin", "", "Origin to proxy to")
	redisURL := flag.String("redis-url", "", "URL to connect to Redis (optional)")
	clearCache := flag.Bool("clear-cache", false, "Clear the cache on start")

	flag.Parse()

	if *origin == "" {
		fmt.Fprintln(os.Stderr, "origin is required")
		os.Exit(1)
	}

	originURL, err := url.Parse(*origin)
	if err != nil {
		fmt.Fprintln(os.Stderr, "origin must be a valid url")
		os.Exit(1)
	}

	var caching cache.Cache
	if *redisURL != "" {
		redisCache, err := cache.NewRedisCache(*redisURL)
		if err != nil {
			log.Fatalf("Error creating redis cache: %v", err)
		}
		caching = redisCache
	} else {
		memoryCache := cache.NewMemoryCache()
		caching = memoryCache
	}

	if *clearCache {
		if err := caching.Clear(context.Background()); err != nil {
			log.Fatalf("Error clearing cache: %v", err)
		}
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", *port),
		Handler: proxy.Handler(originURL, caching),
	}

	log.Printf("Starting proxy server on port %d", *port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

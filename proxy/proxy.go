package proxy

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/edulustosa/caching-proxy/cache"
)

func Handler(origin *url.URL, caching cache.Cache) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originResponse, err := caching.Get(r.Context(), r.URL.String())
		if err == nil {
			for k, v := range originResponse.Headers {
				w.Header()[k] = v
			}
			fmt.Println("X-Cache: HIT")

			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(originResponse.StatusCode)
			w.Write([]byte(originResponse.Body))
			return
		}

		targetURL := origin.String() + r.URL.RequestURI()
		proxyRequest, err := http.NewRequest(r.Method, targetURL, r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		proxyResponse, err := http.DefaultClient.Do(proxyRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer proxyResponse.Body.Close()

		body, err := io.ReadAll(proxyResponse.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		originResponse = cache.OriginResponse{
			StatusCode: proxyResponse.StatusCode,
			Headers:    proxyResponse.Header,
			Body:       string(body),
		}

		err = caching.Set(r.Context(), r.URL.String(), originResponse)
		if err != nil {
			log.Printf(
				"failed to cache origin response err = %s url = %s",
				err.Error(),
				targetURL,
			)
		}

		for k, v := range originResponse.Headers {
			w.Header()[k] = v
		}
		fmt.Println("X-Cache: MISS")
		w.Header().Set("X-Cache", "MISS")
		w.WriteHeader(originResponse.StatusCode)
		w.Write([]byte(originResponse.Body))
	})
}

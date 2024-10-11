# Caching Proxy

A CLI tool that starts a caching proxy server. It forwards requests to the actual server and caches the responses. If the same request is made again, it returns the cached response instead of forwarding the request to the server.

## Usage

``` bash
caching-proxy --port <port> --origin <origin> [--redis-url <redis-url>] [--clear-cache]
```

- `--port`: Specifies the port to run the proxy server. Defaults to 3000.
- `--origin`: The origin URL to which the proxy will forward requests.
- `--redis-url`: An optional redis url to serve as the caching storage. If not specified, the cache will be stored in memory. The url format is `redis://<user>:<pass>@localhost:6379/<db>`.
- `--clear-cache`: An optional flag to clear the cache.

### Example

``` bash
caching-proxy --port 3001 --origin https://jsonplaceholder.typicode.com
```

## Installation

You can download the linux binaries from the [releases page](https://github.com/edulustosa/caching-proxy/releases) or build from source. You must have [Go installed](https://go.dev/doc/install)

``` bash
go build .
```

## Conclusion

This is a challenge feature in <https://roadmap.sh/projects/caching-server> that I used to learn caching with redis. Thanks for checking out!

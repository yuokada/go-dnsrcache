# go-dnsrcache

[![GoDoc](https://godoc.org/github.com/yuokada/go-dnsrcache?status.svg)](https://godoc.org/github.com/yuokada/go-dnsrcache)
[![Golang CI](https://github.com/yuokada/go-dnsrcache/actions/workflows/golang.yml/badge.svg)](https://github.com/yuokada/go-dnsrcache/actions/workflows/golang.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/yuokada/go-dnsrcache)](https://goreportcard.com/report/github.com/yuokada/go-dnsrcache)

dnsrcache caches DNS reverse lookups


inspired by [karlseguin/dnscache](https://github.com/karlseguin/dnscache "karlseguin/dnscache: A DNS Cache for Go")


## Usage

### Installation

To install the package, use:

```bash
go get github.com/yuokada/go-dnsrcache
```

### Creating a DNS Cache

You can create a new DNS cache with a default TTL (Time-To-Live) for cached entries:

```go
package main

import (
    "fmt"
    "github.com/yuokada/go-dnsrcache"
    "time"
)

func main() {
    // Create a DNS cache with a default TTL of 10 seconds
    cache := dnsrcache.NewDNSCache(10 * time.Second)

    // Fetch domains for an IP address
    address := "127.0.0.1"
    domains, err := cache.Fetch(address)
    if err != nil {
        fmt.Printf("Error fetching domains for %s: %v\n", address, err)
        return
    }
    fmt.Printf("Domains for %s: %v\n", address, domains)

    // Close the cache to stop auto-refresh
    cache.Close()
}
```

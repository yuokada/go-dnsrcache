// Package dnscache caches DNS reverse lookups
package dnsrcache

import (
	"testing"
	"time"
	"fmt"
)

var defaultDuration = time.Duration(10 * time.Second)
const(
	ExampleAddr   = "127.0.0.1"
)

var addr_example_com string

func TestNewDNSCache(t *testing.T) {
	s := NewDNSCache(defaultDuration)
	if s.defaultTTL != defaultDuration {
		t.Fatal("duration isn't match!")
	}
}

func TestFetchExamplecom(t *testing.T) {
	cache := NewDNSCache(defaultDuration)
	//a, err:=cache.Fetch(sampleIP)
	a, err:=cache.Fetch(ExampleAddr)
	if err != nil {
		t.Errorf("%#v\n", a)
	}
}

func TestDNSCache_AutoRefresh(t *testing.T) {
	cache  := NewDNSCache(defaultDuration)
	cache.Fetch(ExampleAddr)

	// sleep => cache clear.
	time.Sleep(time.Duration(3 * time.Second))

	cache.Fetch(ExampleAddr)
}

// Example Test

func ExampleDNSCache_Fetch() {
	ttl := time.Duration(10 * time.Millisecond)
	cache := NewDNSCache(ttl)
	hosts, err := cache.Fetch(ExampleAddr)
	if err != nil {
		return
	}
	time.Sleep(time.Duration(100 * time.Millisecond))
	hosts,err = cache.Fetch(ExampleAddr)
	fmt.Println(hosts[0])
	// Output:
	// localhost
}
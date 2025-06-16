// Package dnscache caches DNS reverse lookups
package dnsrcache

import (
	"fmt"
	"testing"
	"time"
)

var defaultDuration = 10 * time.Second

const ExampleAddr = "127.0.0.1"

func TestNewDNSCache(t *testing.T) {
	t.Run("Default TTL is set correctly", func(t *testing.T) {
		cache := NewDNSCache(defaultDuration)
		if cache.defaultTTL != defaultDuration {
			t.Fatalf("expected defaultTTL %v, got %v", defaultDuration, cache.defaultTTL)
		}
	})
}

func TestFetch(t *testing.T) {
	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{"Valid address", ExampleAddr, false},
		{"Invalid address", "invalid-address", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewDNSCache(defaultDuration)
			_, err := cache.Fetch(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDNSCache_AutoRefresh(t *testing.T) {
	t.Run("Auto-refresh clears expired cache", func(t *testing.T) {
		cache := NewDNSCache(3 * time.Second)
		_, err := cache.Fetch(ExampleAddr)
		if err != nil {
			t.Fatalf("Fetch() failed: %v", err)
		}

		// Wait for cache to expire
		time.Sleep(4 * time.Second)

		_, err = cache.Fetch(ExampleAddr)
		if err != nil {
			t.Fatalf("Fetch() failed after auto-refresh: %v", err)
		}
	})
}

// Example Test

func ExampleDNSCache_Fetch() {
	ttl := 10 * time.Millisecond
	cache := NewDNSCache(ttl)
	hosts, err := cache.Fetch(ExampleAddr)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	time.Sleep(100 * time.Millisecond)
	hosts, _ = cache.Fetch(ExampleAddr)
	fmt.Println(hosts[0])
	// Output:
	// localhost
}

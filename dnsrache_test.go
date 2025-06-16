// Package dnscache caches DNS reverse lookups
package dnsrcache

import (
	"fmt"
	"testing"
	"time"
)

var defaultDuration = 10 * time.Second

const ExampleAddr = "127.0.0.1"

func TestNewDNSReverseCache(t *testing.T) {
	t.Run("Default TTL is set correctly", func(t *testing.T) {
		cache := NewDNSReverseCache(defaultDuration)
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
			cache := NewDNSReverseCache(defaultDuration)
			_, err := cache.Fetch(tt.address)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fetch() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDNSReverseCache_AutoRefresh(t *testing.T) {
	t.Run("Auto-refresh clears expired cache", func(t *testing.T) {
		cache := NewDNSReverseCache(3 * time.Second)
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

func ExampleDNSReverseCache_Fetch() {
	cache := NewDNSReverseCache(10 * time.Millisecond)
	hosts, err := cache.Fetch("127.0.0.1")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	time.Sleep(100 * time.Millisecond)
	hosts, _ = cache.Fetch("127.0.0.1")
	fmt.Println(hosts[0])
	// Output:
	// localhost
}

func ExampleDNSReverseCache_Fetch_1_1_1_1() {
	cache := NewDNSReverseCache(10 * time.Millisecond)
	hosts, err := cache.Fetch("1.1.1.1")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	time.Sleep(100 * time.Millisecond)
	hosts, _ = cache.Fetch("1.1.1.1")
	fmt.Println(hosts[0])
	// Output:
	// one.one.one.one.
}

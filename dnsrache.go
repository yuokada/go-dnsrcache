// Package dnsrcache caches DNS reverse lookups
package dnsrcache

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type fqdns struct {
	domains []string
	expires time.Time
}

// DNSCache is Cache struct
type DNSCache struct {
	sync.RWMutex
	defaultTTL time.Duration
	cache      map[string]*fqdns
}

// NewDNSCache : New DNSCache struct with TTL (if TTL <= 0, cache isn't clear)
func NewDNSCache(defaultTTL time.Duration) *DNSCache {
	dcache := &DNSCache{
		defaultTTL: defaultTTL,
		cache:      make(map[string]*fqdns),
	}
	if defaultTTL > 0 {
		go dcache.autoRefresh()
	}
	return dcache
}

// SetTTL : Set a TTL, overwriting the defaultTTL
func (d *DNSCache) SetTTL(ttl time.Duration) error {
	if ttl > 0 {
		d.defaultTTL = ttl
		return nil
	}
	return fmt.Errorf("invalid ttl. ttl wasn't set")
}

// Fetch : Get all of the addresses' ips
func (d *DNSCache) Fetch(address string) ([]string, error) {
	d.RLock()
	value, exists := d.cache[address]
	d.RUnlock()
	if exists {
		now := time.Now()
		if value.expires.After(now) {
			return value.domains, nil
		}
	}

	return d.LookupAddr(address)
}

// LookupAddr : Lookup an address' ip, circumventing the cache
func (d *DNSCache) LookupAddr(address string) ([]string, error) {
	results, err := net.LookupAddr(address)
	if err != nil {
		return nil, err
	}

	expires := time.Now().Add(d.defaultTTL)
	d.Lock()
	d.cache[address] = &fqdns{
		domains: results,
		expires: expires,
	}
	d.Unlock()
	return results, nil
}

// Refresh : Remove expired items (called automatically)
func (d *DNSCache) Refresh() {
	now := time.Now()
	d.RLock()
	for key, value := range d.cache {
		if value.expires.Before(now) {
			delete(d.cache, key)
		}
	}
	d.RUnlock()
}

func (d *DNSCache) autoRefresh() {
	for {
		select {
		case <-time.After(d.defaultTTL):
			// defaultTTLでRefresh()させる
			d.Refresh()
		}
	}
}

// Package dnscache caches DNS reverse lookups
package dnsrcache

import (
	"log"
	"net"
	"sync"
	"time"
)

type fqdns struct {
	domains []string
	expires time.Time
}

type DNSCache struct {
	sync.RWMutex
	defaultTTL time.Duration
	cache      map[string]*fqdns
}

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

// Set a TTL, overwriting the defaultTTL
func (d *DNSCache) SetTTL(ttl time.Duration) {
	if ttl > 0 {
		d.defaultTTL = ttl
	} else {
		log.Println("invalid ttl. ttl wasn't set.")
	}
}

// Get all of the addresses' ips
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

// Lookup an address' ip, circumventing the cache
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

// Remove expired items (called automatically)
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

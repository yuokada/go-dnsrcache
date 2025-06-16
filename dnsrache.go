// Package dnsrcache caches DNS reverse lookups
package dnsrcache

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type fqdns struct {
	domains []string
	expires time.Time
}

// DNSReverseCache is a cache for DNS reverse lookups.
type DNSReverseCache struct {
	sync.RWMutex
	defaultTTL time.Duration
	cache      map[string]*fqdns
	cancel     context.CancelFunc
}

// NewDNSReverseCache creates a new DNSReverseCache with a default TTL. If TTL <= 0, cache isn't cleared automatically.
func NewDNSReverseCache(defaultTTL time.Duration) *DNSReverseCache {
	dcache := &DNSReverseCache{
		defaultTTL: defaultTTL,
		cache:      make(map[string]*fqdns),
	}
	if defaultTTL > 0 {
		ctx, cancel := context.WithCancel(context.Background())
		dcache.cancel = cancel
		go dcache.autoRefresh(ctx)
	}
	return dcache
}

// SetTTL sets a TTL, overwriting the defaultTTL.
func (d *DNSReverseCache) SetTTL(ttl time.Duration) error {
	if ttl > 0 {
		d.defaultTTL = ttl
		return nil
	}
	return fmt.Errorf("invalid ttl. ttl wasn't set")
}

// Fetch returns the cached domains for an address or looks them up if expired/missing.
func (d *DNSReverseCache) Fetch(address string) ([]string, error) {
	d.RLock()
	value, exists := d.cache[address]
	d.RUnlock()
	if exists {
		now := time.Now()
		if value.expires.After(now) {
			return value.domains, nil
		}
	}
	return d.LookupAddr(context.Background(), address)
}

// LookupAddr looks up an address, bypassing the cache.
func (d *DNSReverseCache) LookupAddr(ctx context.Context, address string) ([]string, error) {
	results, err := net.DefaultResolver.LookupAddr(ctx, address)
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

// Refresh removes expired items from the cache.
func (d *DNSReverseCache) Refresh() {
	now := time.Now()
	d.Lock()
	for key, value := range d.cache {
		if value.expires.Before(now) {
			delete(d.cache, key)
		}
	}
	d.Unlock()
}

// autoRefresh periodically calls Refresh at intervals of defaultTTL.
func (d *DNSReverseCache) autoRefresh(ctx context.Context) {
	ticker := time.NewTicker(d.defaultTTL)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.Refresh()
		}
	}
}

// Close stops the auto-refresh goroutine, if running.
func (d *DNSReverseCache) Close() {
	if d.cancel != nil {
		d.cancel()
	}
}

package dnsrcache

import (
	"math/rand"
	"net"
	"sync"
	"time"
)

// value holds cached IPs and their expiration.
type value struct {
	ips     []net.IP
	ipv4s   []net.IP
	expires time.Time
}

// Resolver caches DNS lookups.
type Resolver struct {
	sync.RWMutex
	stop       chan struct{}
	minTTL     time.Duration
	defaultTTL time.Duration
	cache      map[string]*value
	ttls       map[string]time.Duration
}

// New creates a new Resolver with the given default TTL.
func New(defaultTTL time.Duration) *Resolver {
	resolver := &Resolver{
		minTTL:     defaultTTL,
		defaultTTL: defaultTTL,
		stop:       make(chan struct{}),
		cache:      make(map[string]*value),
		ttls:       make(map[string]time.Duration),
	}
	if defaultTTL > 0 {
		go resolver.autoRefresh()
	}
	return resolver
}

// TTL sets a TTL for a specific address, overwriting the defaultTTL.
func (r *Resolver) TTL(address string, ttl time.Duration) {
	r.ttls[address] = ttl
	if ttl < r.minTTL {
		r.minTTL = ttl
	}
}

// Fetch returns all IPs for the address, using the cache if available.
func (r *Resolver) Fetch(address string) ([]net.IP, error) {
	r.RLock()
	value, exists := r.cache[address]
	r.RUnlock()
	if exists {
		return value.ips, nil
	}
	return r.Lookup(address)
}

// FetchOne returns one IP for the address.
func (r *Resolver) FetchOne(address string) (net.IP, error) {
	ips, err := r.Fetch(address)
	if err != nil || len(ips) == 0 {
		return nil, err
	}
	if len(ips) == 1 {
		return ips[0], nil
	}
	return ips[rand.Intn(len(ips))], nil
}

// FetchOneString returns one IP as a string for the address.
func (r *Resolver) FetchOneString(address string) (string, error) {
	ip, err := r.FetchOne(address)
	if err != nil || ip == nil {
		return "", err
	}
	return ip.String(), nil
}

// FetchV4 returns all IPv4 addresses for the address.
func (r *Resolver) FetchV4(address string) ([]net.IP, error) {
	r.RLock()
	value, exists := r.cache[address]
	r.RUnlock()
	if exists {
		return value.ipv4s, nil
	}
	_, err := r.Lookup(address)
	if err != nil {
		return nil, err
	}
	r.RLock()
	value, exists = r.cache[address]
	r.RUnlock()
	if exists {
		return value.ipv4s, nil
	}
	return nil, nil
}

// FetchOneV4 returns one IPv4 address for the address.
func (r *Resolver) FetchOneV4(address string) (net.IP, error) {
	ips, err := r.FetchV4(address)
	if err != nil || len(ips) == 0 {
		return nil, err
	}
	if len(ips) == 1 {
		return ips[0], nil
	}
	return ips[rand.Intn(len(ips))], nil
}

// FetchOneV4String returns one IPv4 address as a string for the address.
func (r *Resolver) FetchOneV4String(address string) (string, error) {
	ip, err := r.FetchOneV4(address)
	if err != nil || ip == nil {
		return "", err
	}
	return ip.String(), nil
}

// Refresh reloads expired items. Called automatically by default.
func (r *Resolver) Refresh() {
	now := time.Now()
	r.RLock()
	addresses := make([]string, 0, len(r.cache))
	for key, value := range r.cache {
		if value.expires.Before(now) {
			addresses = append(addresses, key)
		}
	}
	r.RUnlock()

	for _, address := range addresses {
		r.Lookup(address)
		time.Sleep(10 * time.Millisecond)
	}
}

// Lookup performs a DNS lookup and updates the cache.
func (r *Resolver) Lookup(address string) ([]net.IP, error) {
	ips, err := net.LookupIP(address)
	if err != nil {
		return nil, err
	}

	v4s := make([]net.IP, 0, len(ips))
	for _, ip := range ips {
		if ip.To4() != nil {
			v4s = append(v4s, ip)
		}
	}

	ttl, ok := r.ttls[address]
	if !ok {
		ttl = r.defaultTTL
	}
	expires := time.Now().Add(ttl)
	r.Lock()
	r.cache[address] = &value{
		ips:     ips,
		ipv4s:   v4s,
		expires: expires,
	}
	r.Unlock()
	return ips, nil
}

// Stop stops the background refresher. Once stopped, it cannot be started again.
func (r *Resolver) Stop() {
	close(r.stop)
}

func (r *Resolver) autoRefresh() {
	for {
		select {
		case <-r.stop:
			return
		case <-time.After(r.minTTL):
			r.Refresh()
		}
	}
}

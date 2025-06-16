package dnsrcache

import (
	"errors"
	"net"
	"sort"
	"testing"
	"time"
)

func TestResolver(t *testing.T) {
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{"FetchReturnsAndErrorOnInvalidLookup", testFetchReturnsAndErrorOnInvalidLookup},
		{"FetchReturnsAListOfIps", testFetchReturnsAListOfIps},
		{"Fetchv4ReturnsAListOfIps", testFetchv4ReturnsAListOfIps},
		{"CallingLookupAddsTheItemToTheCache", testCallingLookupAddsTheItemToTheCache},
		{"FetchLoadsValueFromTheCache", testFetchLoadsValueFromTheCache},
		{"FetchOneLoadsAValue", testFetchOneLoadsAValue},
		{"FetchOneStringLoadsAValue", testFetchOneStringLoadsAValue},
		{"FetchLoadsTheIpAndCachesIt", testFetchLoadsTheIpAndCachesIt},
		{"ItReloadsTheIpsAtAGivenInterval", testItReloadsTheIpsAtAGivenInterval},
	}
	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}

func testFetchReturnsAndErrorOnInvalidLookup(t *testing.T) {
	ips, err := New(0).Lookup("invalid.openmymind.io")
	if ips != nil {
		t.Errorf("expected nil, got %v", ips)
	}
	if err == nil || !errors.Is(err, err) { // just check error is not nil
		t.Errorf("expected error, got %v", err)
	}
}

func testFetchReturnsAListOfIps(t *testing.T) {
	ips, _ := New(0).Lookup("go-dnscache.openmymind.io")
	assertIps(t, ips, []string{"8.8.8.8", "8.8.4.4", "2404:6800:4005:8050::1014"})
}

func testFetchv4ReturnsAListOfIps(t *testing.T) {
	ips, _ := New(0).FetchV4("go-dnscache.openmymind.io")
	assertIps(t, ips, []string{"8.8.8.8", "8.8.4.4"})
}

func testCallingLookupAddsTheItemToTheCache(t *testing.T) {
	r := New(0)
	r.Lookup("go-dnscache.openmymind.io")
	assertIps(t, r.cache["go-dnscache.openmymind.io"].ips, []string{"8.8.8.8", "8.8.4.4", "2404:6800:4005:8050::1014"})
}

func testFetchLoadsValueFromTheCache(t *testing.T) {
	r := New(0)
	r.cache["invalid.openmymind.io"] = &value{
		ips:     []net.IP{net.ParseIP("1.1.2.3")},
		ipv4s:   []net.IP{net.ParseIP("1.1.2.3")},
		expires: time.Now(),
	}
	ips, _ := r.Fetch("invalid.openmymind.io")
	assertIps(t, ips, []string{"1.1.2.3"})
}

func testFetchOneLoadsAValue(t *testing.T) {
	r := New(0)
	r.cache["something.openmymind.io"] = &value{
		ips:     []net.IP{net.ParseIP("1.1.2.3"), net.ParseIP("100.100.102.103")},
		ipv4s:   []net.IP{net.ParseIP("1.1.2.3"), net.ParseIP("100.100.102.103")},
		expires: time.Now(),
	}
	ip, _ := r.FetchOne("something.openmymind.io")
	ipStr := ip.String()
	if ipStr != "100.100.102.103" && ipStr != "1.1.2.3" {
		t.Errorf("expected ip to be one of two ips, got %s", ipStr)
	}
}

func testFetchOneStringLoadsAValue(t *testing.T) {
	r := New(0)
	r.cache["something.openmymind.io"] = &value{
		ips:     []net.IP{net.ParseIP("100.100.102.103"), net.ParseIP("100.100.102.104")},
		ipv4s:   []net.IP{net.ParseIP("100.100.102.103"), net.ParseIP("100.100.102.104")},
		expires: time.Now(),
	}
	ip, _ := r.FetchOneString("something.openmymind.io")
	if ip != "100.100.102.103" && ip != "100.100.102.104" {
		t.Errorf("expected ip to be one of two ips, got %s", ip)
	}
}

func testFetchLoadsTheIpAndCachesIt(t *testing.T) {
	r := New(0)
	ips, _ := r.Fetch("go-dnscache.openmymind.io")
	assertIps(t, ips, []string{"8.8.4.4", "8.8.8.8", "2404:6800:4005:8050::1014"})
	assertIps(t, r.cache["go-dnscache.openmymind.io"].ips, []string{"8.8.4.4", "8.8.8.8", "2404:6800:4005:8050::1014"})
}

func testItReloadsTheIpsAtAGivenInterval(t *testing.T) {
	r := New(time.Nanosecond)
	r.cache["go-dnscache.openmymind.io"] = &value{expires: time.Now().Add(-time.Minute)}
	r.Refresh()
	assertIps(t, r.cache["go-dnscache.openmymind.io"].ips, []string{"8.8.4.4", "8.8.8.8", "2404:6800:4005:8050::1014"})
}

func assertIps(t *testing.T, actuals []net.IP, expected []string) {
	t.Helper()
	if len(actuals) != len(expected) {
		t.Fatalf("expected %d IPs, got %d", len(expected), len(actuals))
	}
	ips := make([]string, len(actuals))
	for i, ip := range actuals {
		ips[i] = ip.String()
	}
	sort.Strings(ips)
	sort.Strings(expected)
	for i, ip := range ips {
		if ip != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], ip)
		}
	}
}

package ipcritbit_test

import (
	"net/netip"
	"testing"

	"github.com/gaissmai/ipcritbit"
)

func TestNetip(t *testing.T) {
	rtbl := ipcritbit.New()

	addr4 := netip.MustParseAddr("192.168.1.1")
	host4 := netip.MustParsePrefix("192.168.1.1/32")
	cidr4 := netip.MustParsePrefix("192.168.1.0/24")

	if v, ok := rtbl.Get(cidr4); v != nil || ok {
		t.Errorf("Get() - phantom: %v, %v", v, ok)
	}
	if r, v := rtbl.LookupCIDR(cidr4); r.IsValid() || v != nil {
		t.Errorf("Match() - phantom: %v, %v", r, v)
	}
	if r, v := rtbl.LookupIP(addr4); r.IsValid() || v != nil {
		t.Errorf("MatchIP() - phantom: %v, %v", r, v)
	}
	if v, ok := rtbl.Delete(cidr4); v != nil || ok {
		t.Errorf("Delete() - phantom: %v, %v", v, ok)
	}

	rtbl.Add(cidr4, &cidr4)

	if v, ok := rtbl.Get(cidr4); v != &cidr4 || !ok {
		t.Errorf("Get() - failed: %v, %v", v, ok)
	}
	if r, v := rtbl.LookupCIDR(host4); !r.IsValid() || r != cidr4 || v != &cidr4 {
		t.Errorf("Match() - failed: %v, %v", r, v)
	}
	if r, v := rtbl.LookupIP(addr4); !r.IsValid() || r != cidr4 || v != &cidr4 {
		t.Errorf("MatchIP() - failed: %v, %v", r, v)
	}
	if v, ok := rtbl.Delete(cidr4); v != &cidr4 || !ok {
		t.Errorf("Delete() - failed: %v, %v", v, ok)
	}
}

func checkMatchIP(t *testing.T, trie *ipcritbit.RouteTable, probe, expect string) {
	ip := netip.MustParseAddr(probe)
	route, value := trie.LookupIP(ip)
	if cidr := route.String(); expect != cidr {
		t.Errorf("MatchIP() - %s: expected [%s], actual [%s]", probe, expect, cidr)
	}
	if value == nil {
		t.Errorf("MatchIP() - %s: no value", probe)
	}

	s, ok := value.(string)
	if !ok {
		t.Errorf("MatchIP() - %s: value is not of type string", probe)
	}
	if s != route.String() {
		t.Errorf("MatchIP() - %s: expected [%s], got [%s]", probe, route.String(), s)
	}
}

func buildTestNetip(t *testing.T) *ipcritbit.RouteTable {
	rtbl := ipcritbit.New()

	cidrs := []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("192.168.0.0/16"),
		netip.MustParsePrefix("192.168.1.0/24"),
		netip.MustParsePrefix("192.168.1.0/28"),
		netip.MustParsePrefix("192.168.1.0/32"),
		netip.MustParsePrefix("192.168.1.1/32"),
		netip.MustParsePrefix("192.168.1.2/32"),
		netip.MustParsePrefix("192.168.1.32/27"),
		netip.MustParsePrefix("192.168.1.32/30"),
		netip.MustParsePrefix("192.168.2.1/32"),
		netip.MustParsePrefix("192.168.2.2/32"),

		netip.MustParsePrefix("2001:db8::/32"),
		netip.MustParsePrefix("2001:db8::/64"),
		netip.MustParsePrefix("fe80::/10"),
		netip.MustParsePrefix("::/0"),
	}

	for _, cidr := range cidrs {
		rtbl.Add(cidr, cidr.String())
	}
	return rtbl
}

func TestNetipMatchIP(t *testing.T) {
	rtbl := buildTestNetip(t)

	checkMatchIP(t, rtbl, "10.0.0.0", "10.0.0.0/8")
	checkMatchIP(t, rtbl, "192.168.1.0", "192.168.1.0/32")
	checkMatchIP(t, rtbl, "192.168.1.3", "192.168.1.0/28")
	checkMatchIP(t, rtbl, "192.168.1.128", "192.168.1.0/24")
	checkMatchIP(t, rtbl, "192.168.2.128", "192.168.0.0/16")
	checkMatchIP(t, rtbl, "192.168.1.1", "192.168.1.1/32")
	checkMatchIP(t, rtbl, "192.168.1.2", "192.168.1.2/32")
	checkMatchIP(t, rtbl, "192.168.1.3", "192.168.1.0/28")
	checkMatchIP(t, rtbl, "192.168.1.32", "192.168.1.32/30")
	checkMatchIP(t, rtbl, "192.168.1.35", "192.168.1.32/30")
	checkMatchIP(t, rtbl, "192.168.1.36", "192.168.1.32/27")
	checkMatchIP(t, rtbl, "192.168.1.63", "192.168.1.32/27")
	checkMatchIP(t, rtbl, "192.168.1.64", "192.168.1.0/24")
	checkMatchIP(t, rtbl, "192.168.2.2", "192.168.2.2/32")
	checkMatchIP(t, rtbl, "192.168.2.3", "192.168.0.0/16")
	checkMatchIP(t, rtbl, "2001:db8:0:0::", "2001:db8::/64")
	checkMatchIP(t, rtbl, "2001:db8:0:1::", "2001:db8::/32")
	checkMatchIP(t, rtbl, "fe80::1", "fe80::/10")
	checkMatchIP(t, rtbl, "dead:beef::ffff", "::/0")
}

func TestNetipWalk(t *testing.T) {
	rtbl := buildTestNetip(t)

	var c int
	f := func(p netip.Prefix, v interface{}) bool {
		c += 1
		return true
	}

	c = 0
	rtbl.Walk(netip.Prefix{}, f)
	if c != 15 {
		t.Errorf("Walk() - %d: full walk", c)
	}

	p := netip.MustParsePrefix("192.168.1.1/32")
	c = 0
	rtbl.Walk(p, f)
	if c != 6 {
		t.Errorf("Walk() - %d: has start route v4", c)
	}

	p = netip.MustParsePrefix("2001:db8::/64")
	c = 0
	rtbl.Walk(p, f)
	if c != 2 {
		t.Errorf("Walk() - %d: has start route v6", c)
	}

	p = netip.MustParsePrefix("10.0.0.0/0")
	c = 0
	rtbl.Walk(p, f)
	if c != 0 {
		t.Errorf("Walk() - %d: not found start route v4", c)
	}

	p = netip.MustParsePrefix("dead:beef::ffee:aabb/96")
	c = 0
	rtbl.Walk(p, f)
	if c != 0 {
		t.Errorf("Walk() - %d: not found start route v4", c)
	}
}

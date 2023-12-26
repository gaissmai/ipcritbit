package ipcritbit_test

import (
	"math/rand"
	"net/netip"
	"testing"

	"github.com/gaissmai/ipcritbit"
)

var (
	routeCount2 int = 1_000_000
	cidrs       []netip.Prefix
)

func init() {
	cidrs = make([]netip.Prefix, routeCount2)
	random := rand.New(rand.NewSource(0))
	for i := 0; i < len(cidrs); i++ {
		cidrs[i] = genCIDR(random)
	}
}

func genCIDR(rand *rand.Rand) netip.Prefix {
	ip := rand.Int31()
	bits := rand.Intn(33)

	a4 := [4]byte{byte(ip >> 24), byte(ip >> 16), byte(ip >> 8), byte(ip)}
	addr := netip.AddrFrom4(a4)

	return netip.PrefixFrom(addr, bits)
}

func buildRTable(keys []netip.Prefix) *ipcritbit.RouteTable {
	rtbl := ipcritbit.New()
	for i := 0; i < len(keys); i++ {
		rtbl.Add(keys[i], nil)
	}
	return rtbl
}

func BenchmarkNetipBuild(b *testing.B) {
	for i := 0; i < b.N; i++ {
		buildRTable(cidrs)
	}
}

/*
func BenchmarkNetipGet(b *testing.B) {
	rtbl := buildRTable(cidrs)
	random := rand.New(rand.NewSource(0))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		k := cidrs[random.Intn(routeCount2)]
		rtbl.Get(k)
	}
}

func BenchmarkNetipDelete(b *testing.B) {
	rtbl := buildRTable(cidrs)
	random := rand.New(rand.NewSource(0))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		k := cidrs[random.Intn(keyCount)]
		rtbl.Delete(k)
	}
}

func BenchmarkNetipMatch(b *testing.B) {
	rtbl := buildRTable(cidrs)
	random := rand.New(rand.NewSource(0))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := genCIDR(random)
		rtbl.Match(s)
	}
}
*/

package ipcritbit

import (
	"io"
	"net/netip"
)

// IP routing table.
type RouteTable struct {
	tree4 *critBitTree
	tree6 *critBitTree
}

// Create IP routing table
func New() *RouteTable {
	return &RouteTable{
		tree4: newTree(),
		tree6: newTree(),
	}
}

// Add a route.
func (t *RouteTable) Add(p netip.Prefix, value interface{}) {
	key := pfxToKey(p)
	if p.Addr().Is4() {
		t.tree4.set(key, value)
		return
	}
	t.tree6.set(key, value)
}

// Delete a specific route.
func (t *RouteTable) Delete(p netip.Prefix) (value interface{}, ok bool) {
	if p.Addr().Is4() {
		return t.tree4.delete(pfxToKey(p))
	}
	return t.tree6.delete(pfxToKey(p))
}

// Get a specific route.
func (t *RouteTable) Get(p netip.Prefix) (value interface{}, ok bool) {
	if p.Addr().Is4() {
		return t.tree4.get(pfxToKey(p))
	}
	return t.tree6.get(pfxToKey(p))
}

// Return a specific route by using the longest prefix matching.
func (t *RouteTable) LookupCIDR(p netip.Prefix) (route netip.Prefix, value interface{}) {
	if p.Addr().Is4() {
		if k, v := t.match4(pfxToKey(p)); k != nil {
			unmarshal(&route, k)
			value = v
		}
		return
	}
	if k, v := t.match6(pfxToKey(p)); k != nil {
		unmarshal(&route, k)
		value = v
	}
	return
}

// Return a specific route by using the longest prefix matching.
func (t *RouteTable) LookupIP(ip netip.Addr) (route netip.Prefix, value interface{}) {
	k, v := t.matchIP(ip)
	if k != nil {
		unmarshal(&route, k)
		value = v
	}
	return
}

func (t *RouteTable) matchIP(ip netip.Addr) (k []byte, v interface{}) {
	if ip.Is4() {
		p := netip.PrefixFrom(ip, 32)
		k, v = t.match4(pfxToKey(p))
		return
	}
	p := netip.PrefixFrom(ip, 128)
	k, v = t.match6(pfxToKey(p))
	return
}

func (t *RouteTable) match4(key []byte) ([]byte, interface{}) {
	if t.tree4.items > 0 {
		if node := lookup(&t.tree4.root, key, false); node != nil {
			return node.external.key, node.external.value
		}
	}
	return nil, nil
}

func (t *RouteTable) match6(key []byte) ([]byte, interface{}) {
	if t.tree6.items > 0 {
		if node := lookup(&t.tree6.root, key, false); node != nil {
			return node.external.key, node.external.value
		}
	}
	return nil, nil
}

func lookup(p *node, key []byte, backtracking bool) *node {
	if p.internal != nil {
		var direction int
		if p.internal.offset == len(key)-1 {
			// selecting the larger side when comparing the mask
			direction = 1
		} else if backtracking {
			direction = 0
		} else {
			direction = p.internal.direction(key)
		}

		if c := lookup(&p.internal.child[direction], key, backtracking); c != nil {
			return c
		}
		if direction == 1 {
			// search other node
			return lookup(&p.internal.child[0], key, true)
		}
		return nil
	} else {
		nlen := len(p.external.key)
		if nlen != len(key) {
			return nil
		}

		// check mask
		mask := p.external.key[nlen-1]
		if mask > key[nlen-1] {
			return nil
		}

		// compare both keys with mask
		div := int(mask >> 3)
		for i := 0; i < div; i++ {
			if p.external.key[i] != key[i] {
				return nil
			}
		}
		if mod := uint(mask & 0x07); mod > 0 {
			bit := 8 - mod
			if p.external.key[div] != key[div]&(0xff>>bit<<bit) {
				return nil
			}
		}
		return p
	}
}

// Walk iterates routes from a given route, adapter for ipcritbit.walk
// handle is called with arguments route and value (if handle returns `false`, the iteration is aborted)
func (t *RouteTable) Walk(p netip.Prefix, handle func(netip.Prefix, interface{}) bool) {
	key := pfxToKey(p)
	if key == nil {
		t.tree4.walk(nil, func(currentKey []byte, value interface{}) bool {
			return handle(keyToPfx(currentKey), value)
		})

		t.tree6.walk(nil, func(currentKey []byte, value interface{}) bool {
			return handle(keyToPfx(currentKey), value)
		})
		return
	}

	if p.Addr().Is4() {
		t.tree4.walk(key, func(currentKey []byte, value interface{}) bool {
			return handle(keyToPfx(currentKey), value)
		})

		return
	}

	t.tree6.walk(key, func(currentKey []byte, value interface{}) bool {
		return handle(keyToPfx(currentKey), value)
	})

	return
}

// Dump routing table. (for debugging)
func (t *RouteTable) Dump(w io.Writer) {
	t.tree4.dump(w)
	t.tree6.dump(w)
}

// Deletes all routes.
func (t *RouteTable) Clear() {
	t.tree4.clear()
	t.tree6.clear()
}

// Returns number of routes, all IP versions.
func (t *RouteTable) Size() int {
	return t.tree4.items + t.tree6.items
}

// helpers, convert between keys []byte and netip.Prefix

func pfxToKey(p netip.Prefix) []byte {
	if !p.IsValid() {
		return nil
	}

	b, err := p.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return b
}

func keyToPfx(key []byte) netip.Prefix {
	p := netip.Prefix{}
	unmarshal(&p, key)
	return p
}

// ummarshal does not allocate.
func unmarshal(p *netip.Prefix, b []byte) {
	err := p.UnmarshalBinary(b)
	if err != nil {
		panic(err)
	}
}


 **ATTENTION** still in alpha stage!
====================================

ipcritbit
=========
A [critbit-tree](http://cr.yp.to/critbit.html) implementation in golang for fast IP lookup.

The [original](https://github.com/k-sone/critbitgo) has been forked, modified and reduced to LPM-lookups for `net/netip` addresses and prefixes.
Both IP versions are supported transparently. 

Usage
-----

```go
import "github.com/gaissmai/ipcritbit"

type RouteTable struct { // Has unexported fields.  }

func New() RouteTable

func (t RouteTable) Add(p netip.Prefix, value interface{})
func (t RouteTable) Get(p netip.Prefix) (value interface{}, ok bool)
func (t RouteTable) Delete(p netip.Prefix) (value interface{}, ok bool)

func (t RouteTable) LookupIP(ip netip.Addr) (route netip.Prefix, value interface{})
func (t RouteTable) LookupCIDR(p netip.Prefix) (route netip.Prefix, value interface{})

func (t RouteTable) Clear()
func (t RouteTable) Size() int

func (t RouteTable) Walk(callback func(prefix netip.Prefix, value interface{}) bool)
func (t RouteTable) Dump(w io.Writer)
```

License
-------

MIT

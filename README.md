ipcritbit
=========
A [critbit-tree](http://cr.yp.to/critbit.html) implementation in golang for fast IP lookup.

The [original](https://github.com/k-sone/critbitgo) has been forked, modified and reduced to LPM-lookups for `net/netip` addresses and prefixes.
Both IP versions are supported transparently. 

Usage
--------

```go
// New routing table.
rtbl := ipcritbit.New()

 ... TODO
```

License
-------

MIT

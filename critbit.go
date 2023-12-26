package ipcritbit

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
)

// The matrix of most significant bit
var msbMatrix [256]byte

func buildMsbMatrix() {
	for i := 0; i < len(msbMatrix); i++ {
		b := byte(i)
		b |= b >> 1
		b |= b >> 2
		b |= b >> 4
		msbMatrix[i] = b &^ (b >> 1)
	}
}

func init() {
	buildMsbMatrix()
}

type node struct {
	internal *internal
	external *external
}

type internal struct {
	child  [2]node
	offset int
	bit    byte
	cont   bool // if true, key of child[1] contains key of child[0]
}

type external struct {
	key   []byte
	value interface{}
}

// critBitTree
type critBitTree struct {
	root  node
	items int
}

// create a tree.
func newTree() *critBitTree {
	return &critBitTree{}
}

// finding the critical bit.
func (n *external) criticalBit(key []byte) (offset int, bit byte, cont bool) {
	nlen := len(n.key)
	klen := len(key)
	mlen := nlen
	if nlen > klen {
		mlen = klen
	}

	// find first differing byte and bit
	for offset = 0; offset < mlen; offset++ {
		if a, b := key[offset], n.key[offset]; a != b {
			bit = msbMatrix[a^b]
			return
		}
	}

	if nlen < klen {
		bit = msbMatrix[key[offset]]
	} else if nlen > klen {
		bit = msbMatrix[n.key[offset]]
	} else {
		// two keys are equal
		offset = -1
	}
	return offset, bit, true
}

// calculate direction.
func (n *internal) direction(key []byte) int {
	if n.offset < len(key) && (key[n.offset]&n.bit != 0 || n.cont) {
		return 1
	}
	return 0
}

// searching the tree.
func (t *critBitTree) search(key []byte) *node {
	n := &t.root
	for n.internal != nil {
		n = &n.internal.child[n.internal.direction(key)]
	}
	return n
}

// membership testing.
func (t *critBitTree) contains(key []byte) bool {
	if n := t.search(key); n.external != nil && bytes.Equal(n.external.key, key) {
		return true
	}
	return false
}

// get member.
// if `key` is in Trie, `ok` is true.
func (t *critBitTree) get(key []byte) (value interface{}, ok bool) {
	if n := t.search(key); n.external != nil && bytes.Equal(n.external.key, key) {
		return n.external.value, true
	}
	return
}

// insertHelper into the tree (replaceable).
func (t *critBitTree) insertHelper(key []byte, value interface{}, replace bool) bool {
	// an empty tree
	if t.items == 0 {
		t.root.external = &external{
			key:   key,
			value: value,
		}
		t.items = 1
		return true
	}

	n := t.search(key)
	newOffset, newBit, newCont := n.external.criticalBit(key)

	// already exists in the tree
	if newOffset == -1 {
		if replace {
			n.external.value = value
			return true
		}
		return false
	}

	// allocate new node
	newNode := &internal{
		offset: newOffset,
		bit:    newBit,
		cont:   newCont,
	}
	direction := newNode.direction(key)
	newNode.child[direction].external = &external{
		key:   key,
		value: value,
	}

	// insert new node
	wherep := &t.root
	for in := wherep.internal; in != nil; in = wherep.internal {
		if in.offset > newOffset || (in.offset == newOffset && in.bit < newBit) {
			break
		}
		wherep = &in.child[in.direction(key)]
	}

	if wherep.internal != nil {
		newNode.child[1-direction].internal = wherep.internal
	} else {
		newNode.child[1-direction].external = wherep.external
		wherep.external = nil
	}
	wherep.internal = newNode
	t.items += 1
	return true
}

// insert into the tree.
// if `key` is alredy in Trie, return false.
func (t *critBitTree) insert(key []byte, value interface{}) bool {
	return t.insertHelper(key, value, false)
}

// set into the tree.
func (t *critBitTree) set(key []byte, value interface{}) {
	t.insertHelper(key, value, true)
}

// deleting elements.
// if `key` is in Trie, `ok` is true.
func (t *critBitTree) delete(key []byte) (value interface{}, ok bool) {
	// an empty tree
	if t.items == 0 {
		return
	}

	var direction int
	var whereq *node // pointer to the grandparent
	var wherep *node = &t.root

	// finding the best candidate to delete
	for in := wherep.internal; in != nil; in = wherep.internal {
		direction = in.direction(key)
		whereq = wherep
		wherep = &in.child[direction]
	}

	// checking that we have the right element
	if !bytes.Equal(wherep.external.key, key) {
		return
	}
	value = wherep.external.value
	ok = true

	// removing the node
	if whereq == nil {
		wherep.external = nil
	} else {
		othern := whereq.internal.child[1-direction]
		whereq.internal = othern.internal
		whereq.external = othern.external
	}
	t.items -= 1
	return
}

// clearing a tree.
func (t *critBitTree) clear() {
	t.root.internal = nil
	t.root.external = nil
	t.items = 0
}

// return the number of key in a tree.
func (t *critBitTree) size() int {
	return t.items
}

// Search for the longest matching key from the beginning of the given key.
// if `key` is in Trie, `ok` is true.
func (t *critBitTree) longestPrefix(given []byte) (key []byte, value interface{}, ok bool) {
	// an empty tree
	if t.items == 0 {
		return
	}
	return longestPrefixHelper(&t.root, given)
}

func longestPrefixHelper(n *node, key []byte) ([]byte, interface{}, bool) {
	if n.internal != nil {
		direction := n.internal.direction(key)
		if k, v, ok := longestPrefixHelper(&n.internal.child[direction], key); ok {
			return k, v, ok
		}
		if direction == 1 {
			return longestPrefixHelper(&n.internal.child[0], key)
		}
	} else {
		if bytes.HasPrefix(key, n.external.key) {
			return n.external.key, n.external.value, true
		}
	}
	return nil, nil, false
}

// Iterating elements from a given start key.
// handle is called with arguments key and value (if handle returns `false`, the iteration is aborted)
func (t *critBitTree) walk(start []byte, handle func(key []byte, value interface{}) bool) bool {
	if t.items == 0 {
		return true
	}
	var seek bool
	if start != nil {
		seek = true
	}
	return walkHelper(&t.root, start, &seek, handle)
}

func walkHelper(n *node, key []byte, seek *bool, handle func([]byte, interface{}) bool) bool {
	if n.internal != nil {
		var direction int
		if *seek {
			direction = n.internal.direction(key)
		}
		if !walkHelper(&n.internal.child[direction], key, seek, handle) {
			return false
		}
		if !(*seek) && direction == 0 {
			// iteration another side
			return walkHelper(&n.internal.child[1], key, seek, handle)
		}
		return true
	} else {
		if *seek {
			if bytes.Equal(n.external.key, key) {
				// seek completed
				*seek = false
			} else {
				// key is not in Trie
				return false
			}
		}
		return handle(n.external.key, n.external.value)
	}
}

// dump tree. (for debugging)
func (t *critBitTree) dump(w io.Writer) {
	if t.root.internal == nil && t.root.external == nil {
		return
	}
	if w == nil {
		w = os.Stdout
	}
	dumpHelper(w, &t.root, true, "")
}

func dumpHelper(w io.Writer, n *node, right bool, prefix string) {
	var ownprefix string
	if right {
		ownprefix = prefix
	} else {
		ownprefix = prefix[:len(prefix)-1] + "`"
	}

	if in := n.internal; in != nil {
		fmt.Fprintf(w, "%s-- off=%d, bit=%08b(%02x), cont=%v\n", ownprefix, in.offset, in.bit, in.bit, in.cont)
		for i := 0; i < 2; i++ {
			var nextprefix string
			switch i {
			case 0:
				nextprefix = prefix + " |"
				right = true
			case 1:
				nextprefix = prefix + "  "
				right = false
			}
			dumpHelper(w, &in.child[i], right, nextprefix)
		}
	} else {
		fmt.Fprintf(w, "%s-- key=%d (%s)\n", ownprefix, n.external.key, key2str(n.external.key))
	}
	return
}

func key2str(key []byte) string {
	for _, c := range key {
		if !strconv.IsPrint(rune(c)) {
			return hex.EncodeToString(key)
		}
	}
	return string(key)
}

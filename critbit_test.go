package ipcritbit

import (
	"bytes"
	"math/rand"
	"testing"
)

func buildTrie(t *testing.T, keys []string) *critBitTree {
	trie := newTree()
	for _, key := range keys {
		if !trie.insert([]byte(key), key) {
			t.Errorf("insert() - failed insert \"%s\"\n%s", key, dumpTrie(trie))
		}
	}
	return trie
}

func dumpTrie(trie *critBitTree) string {
	buf := bytes.NewBufferString("")
	trie.dump(buf)
	return buf.String()
}

func TestInsert(t *testing.T) {
	// normal build
	keys := []string{"", "a", "aa", "b", "bb", "ab", "ba", "aba", "bab"}
	trie := buildTrie(t, keys)
	dump := dumpTrie(trie)

	// random build
	random := rand.New(rand.NewSource(0))
	for i := 0; i < 10; i++ {
		// shuffle keys
		lkeys := make([]string, len(keys))
		for j, index := range random.Perm(len(keys)) {
			lkeys[j] = keys[index]
		}

		ltrie := buildTrie(t, lkeys)
		ldump := dumpTrie(ltrie)
		if dump != ldump {
			t.Errorf("insert() - different tries\norigin:\n%s\nother:\n%s\n", dump, ldump)
		}
	}

	// error check
	if trie.insert([]byte("a"), nil) {
		t.Error("insert() - check exists")
	}
	if !trie.insert([]byte("c"), nil) {
		t.Error("insert() - check not exists")
	}
}

func TestSet(t *testing.T) {
	keys := []string{"", "a", "aa", "b", "bb", "ab", "ba", "aba", "bab"}
	trie := buildTrie(t, keys)

	trie.set([]byte("a"), 100)
	v, _ := trie.get([]byte("a"))
	if n, ok := v.(int); !ok || n != 100 {
		t.Errorf("set() - failed replace - %v", v)
	}
}

func TestContains(t *testing.T) {
	keys := []string{"", "a", "aa", "b", "bb", "ab", "ba", "aba", "bab"}
	trie := buildTrie(t, keys)

	for _, key := range keys {
		if !trie.contains([]byte(key)) {
			t.Errorf("contains() - not found - %s", key)
		}
	}

	if trie.contains([]byte("aaa")) {
		t.Error("contains() - phantom found")
	}
}

func TestGet(t *testing.T) {
	keys := []string{"", "a", "aa", "b", "bb", "ab", "ba", "aba", "bab"}
	trie := buildTrie(t, keys)

	for _, key := range keys {
		if value, ok := trie.get([]byte(key)); value != key || !ok {
			t.Errorf("Get() - not found - %s", key)
		}
	}

	if value, ok := trie.get([]byte("aaa")); value != nil || ok {
		t.Error("Get() - phantom found")
	}
}

func TestDelete(t *testing.T) {
	keys := []string{"", "a", "aa", "b", "bb", "ab", "ba", "aba", "bab"}
	trie := buildTrie(t, keys)

	for i, key := range keys {
		if !trie.contains([]byte(key)) {
			t.Errorf("Delete() - not exists - %s", key)
		}
		if v, ok := trie.delete([]byte(key)); !ok || v != key {
			t.Errorf("Delete() - failed - %s", key)
		}
		if trie.contains([]byte(key)) {
			t.Errorf("Delete() - exists - %s", key)
		}
		if i != len(keys) {
			for _, key2 := range keys[i+1:] {
				if !trie.contains([]byte(key2)) {
					t.Errorf("Delete() - other not exists - %s", key2)
				}
			}
		}
	}
}

func TestSize(t *testing.T) {
	keys := []string{"", "a", "aa", "b", "bb", "ab", "ba", "aba", "bab"}
	trie := buildTrie(t, keys)
	klen := len(keys)
	if s := trie.size(); s != klen {
		t.Errorf("Size() - expected [%d], actual [%d]", klen, s)
	}

	for i, key := range keys {
		trie.delete([]byte(key))
		if s := trie.size(); s != klen-(i+1) {
			t.Errorf("Size() - expected [%d], actual [%d]", klen, s)
		}
	}
}

func TestLongestPrefix(t *testing.T) {
	keys := []string{"a", "aa", "b", "bb", "ab", "ba", "aba", "bab"}
	trie := buildTrie(t, keys)

	expects := map[string]string{
		"a":   "a",
		"a^":  "a",
		"aaa": "aa",
		"abc": "ab",
		"bac": "ba",
		"bbb": "bb",
		"bc":  "b",
	}
	for g, k := range expects {
		if key, value, ok := trie.longestPrefix([]byte(g)); !ok || string(key) != k || value != k {
			t.Errorf("longestPrefix() - invalid result - %s not %s", key, g)
		}
	}

	if _, _, ok := trie.longestPrefix([]byte{}); ok {
		t.Error("longestPrefix() - invalid result - not empty")
	}
	if _, _, ok := trie.longestPrefix([]byte("^")); ok {
		t.Error("longestPrefix() - invalid result - not empty")
	}
	if _, _, ok := trie.longestPrefix([]byte("c")); ok {
		t.Error("longestPrefix() - invalid result - not empty")
	}
}

func TestWalk(t *testing.T) {
	keys := []string{"", "a", "aa", "b", "bb", "ab", "ba", "aba", "bab"}
	trie := buildTrie(t, keys)

	elems := make([]string, 0, len(keys))
	handle := func(key []byte, value interface{}) bool {
		if k := string(key); k == value {
			elems = append(elems, k)
		}
		return true
	}
	if !trie.walk([]byte{}, handle) {
		t.Error("walk() - invalid result")
	}
	if len(elems) != 9 {
		t.Errorf("walk() - invalid elems length [%v]", elems)
	}
	for i, key := range []string{"", "a", "aa", "ab", "aba", "b", "ba", "bab", "bb"} {
		if key != elems[i] {
			t.Errorf("walk() - not found [%s]", key)
		}
	}

	elems = make([]string, 0, len(keys))
	if !trie.walk([]byte("ab"), handle) {
		t.Error("walk() - invalid result")
	}
	if len(elems) != 6 {
		t.Errorf("walk() - invalid elems length [%v]", elems)
	}
	for i, key := range []string{"ab", "aba", "b", "ba", "bab", "bb"} {
		if key != elems[i] {
			t.Errorf("walk() - not found [%s]", key)
		}
	}

	elems = make([]string, 0, len(keys))
	if !trie.walk(nil, handle) {
		t.Error("walk() - invalid result")
	}
	if len(elems) != 9 {
		t.Errorf("walk() - invalid elems length [%v]", elems)
	}
	for i, key := range []string{"", "a", "aa", "ab", "aba", "b", "ba", "bab", "bb"} {
		if key != elems[i] {
			t.Errorf("walk() - not found [%s]", key)
		}
	}

	elems = make([]string, 0, len(keys))
	handle = func(key []byte, value interface{}) bool {
		if k := string(key); k == value {
			elems = append(elems, k)
		}
		if string(key) == "aa" {
			return false
		}
		return true
	}
	if trie.walk([]byte("a"), handle) {
		t.Error("walk() - invalid result")
	}
	if len(elems) != 2 {
		t.Errorf("walk() - invalid elems length [%v]", elems)
	}
	for i, key := range []string{"a", "aa"} {
		if key != elems[i] {
			t.Errorf("walk() - not found [%s]", key)
		}
	}

	if trie.walk([]byte("^"), handle) {
		t.Error("walk() - invalid result")
	}
	if trie.walk([]byte("aaa"), handle) {
		t.Error("walk() - invalid result")
	}
	if trie.walk([]byte("c"), handle) {
		t.Error("walk() - invalid result")
	}
}

func TestEmptyTree(t *testing.T) {
	trie := newTree()
	key := []byte{0, 1, 2}
	handle := func(_ []byte, _ interface{}) bool { return true }
	assert := func(n string, f func()) {
		defer func() {
			if e := recover(); e != nil {
				t.Errorf("%s() - empty tree : %v", n, e)
			}
		}()
		f()
	}

	assert("contains", func() { trie.contains(key) })
	assert("get", func() { trie.get(key) })
	assert("delete", func() { trie.delete(key) })
	assert("longestPrefix", func() { trie.longestPrefix(key) })
	assert("walk", func() { trie.walk(key, handle) })
}

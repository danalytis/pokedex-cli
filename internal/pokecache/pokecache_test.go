package pokecache

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddGet(t *testing.T) {
	const interval = 5 * time.Second
	cases := []struct {
		key string
		val []byte
	}{
		{
			key: "https://example.com",
			val: []byte("testdata"),
		},
		{
			key: "https://example.com/path",
			val: []byte("moretestdata"),
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			cache := NewCache(interval)
			cache.Add(c.key, c.val)
			val, ok := cache.Get(c.key)
			assert.True(t, ok, "expected to find key")
			assert.Equal(t, string(c.val), string(val))
		})
	}
}

func TestReapLoop(t *testing.T) {
	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime + 5*time.Millisecond
	cache := NewCache(baseTime)
	cache.Add("https://example.com", []byte("testdata"))

	_, ok := cache.Get("https://example.com")
	assert.True(t, ok, "expected to find key")

	time.Sleep(waitTime)

	_, ok = cache.Get("https://example.com")
	assert.False(t, ok, "expected to not find key after expiration")
}

func TestGet_NonExistentKey(t *testing.T) {
	cache := NewCache(5 * time.Millisecond)
	_, ok := cache.Get("pikachu")
	assert.False(t, ok)
}

func TestAdd_OverwriteExistingKey(t *testing.T) {

	cache := NewCache(5 * time.Minute)

	cache.Add("pokemon", []byte("pikachu"))
	cache.Add("pokemon", []byte("charizard"))

	val, ok := cache.Get("pokemon")

	assert.True(t, ok)
	assert.Equal(t, "charizard", string(val))
}

func TestAdd_EmptyKey(t *testing.T) {
	cache := NewCache(5 * time.Minute)
	cache.Add("", []byte("pikachu"))
	_, ok := cache.Get("")
	assert.False(t, ok)
}

func TestAdd_NilValue(t *testing.T) {
	// Test adding with nil byte slice
	cache := NewCache(5 * time.Minute)
	cache.Add("pokemon", nil)
	_, ok := cache.Get("pokemon")
	assert.False(t, ok)
}

func TestAdd_EmptyValue(t *testing.T) {
	cache := NewCache(5 * time.Minute)
	cache.Add("pokemon", []byte(""))
	_, ok := cache.Get("pokemon")
	assert.False(t, ok)
}

// TODO: implement
// Timing and expiration
func TestReapLoop_MultipleEntries(t *testing.T) {
	// Test that reap loop cleans multiple expired entries
}

func TestReapLoop_PartialExpiration(t *testing.T) {
	// Test some entries expire while others don't
}

func TestCache_NoExpiration(t *testing.T) {
	// Test cache with very long interval doesn't expire
}

// Concurrency (advanced)
func TestCache_ConcurrentAccess(t *testing.T) {
	// Test multiple goroutines adding/getting simultaneously
}

func TestCache_ConcurrentReap(t *testing.T) {
	// Test reaping while other operations are happening
}

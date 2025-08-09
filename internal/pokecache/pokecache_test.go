package pokecache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddGet(t *testing.T) {
	const interval = 5 * time.Millisecond
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

// Test that reap loop cleans multiple expired entries
func TestReapLoop_MultipleEntries(t *testing.T) {
	cases := []string{
		"https://example.com",
		"https://example2.com",
		"https://example3.com",
	}

	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime + 5*time.Millisecond
	cache := NewCache(baseTime)
	for _, url := range cases {
		cache.Add(url, []byte("testdata"))
	}

	_, ok := cache.Get("https://example.com")
	assert.True(t, ok, "expected to find key")

	time.Sleep(waitTime)

	for _, url := range cases {
		_, ok = cache.Get(url)
		assert.False(t, ok, "expected to not find key after expiration")
	}
}

// Test some entries expire while others don't
func TestReapLoop_PartialExpiration(t *testing.T) {
	cases := []struct {
		key   string
		found bool
	}{
		{
			key:   "https://example.com",
			found: false,
		},
		{
			key:   "https://example2.com",
			found: false,
		},
		{
			key:   "https://example3.com",
			found: true,
		},
	}
	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime + 5*time.Millisecond
	cache := NewCache(baseTime)

	cache.Add("https://example.com", []byte("testdata"))
	cache.Add("https://example2.com", []byte("testdata"))
	time.Sleep(waitTime)
	cache.Add("https://example3.com", []byte("testdata"))

	for _, c := range cases {
		_, ok := cache.Get(c.key)
		assert.Equal(t, c.found, ok)
	}
}

func TestCache_NoPrematureExpiration(t *testing.T) {
	const baseTime = 1 * time.Second
	const waitTime = 500 * time.Millisecond

	cache := NewCache(baseTime)
	cache.Add("https://example.com", []byte("testdata"))
	time.Sleep(waitTime)

	_, ok := cache.Get("https://example.com")
	assert.True(t, ok, "expected to find key")
}

// Test multiple goroutines adding/getting simultaneously
func TestCache_ConcurrentAccess(t *testing.T) {
	const baseTime = 1 * time.Second

	cache := NewCache(baseTime)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		cache.Add("https://example.com", []byte("testdata"))
	}()

	go func() {
		defer wg.Done()
		cache.Add("https://example2.com", []byte("testdata"))
	}()

	wg.Wait()

	_, ok := cache.Get("https://example.com")
	assert.True(t, ok, "expected to find key")
	_, ok = cache.Get("https://example2.com")
	assert.True(t, ok, "expected to find key")
}

// Test reaping while other operations are happening
func TestCache_ConcurrentReap(t *testing.T) {
	const baseTime = 10 * time.Millisecond
	cache := NewCache(baseTime)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			key := fmt.Sprintf("url%d", i)
			cache.Add(key, []byte("data"))
			time.Sleep(2 * time.Millisecond)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			key := fmt.Sprintf("url%d", i)
			cache.Get(key)
			time.Sleep(2 * time.Millisecond)
		}
	}()

	wg.Wait()
	assert.True(t, true, "concurrent operations completed without panic")
}

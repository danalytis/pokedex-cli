package pokeapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danalytis/pokedexcli/internal/pokecache"
	"github.com/stretchr/testify/assert"
)

func TestFetchAndCache_CacheHit(t *testing.T) {
	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClient(&cache)

	testData := `{"name": "test"}`
	cache.Add("test-url", []byte(testData))

	var result map[string]interface{}
	err := client.fetchAndCache("test-url", &result)

	assert.NoError(t, err)
	assert.Equal(t, "test", result["name"])
}

func TestFetchAndCache_CacheMiss(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"name": "test"}`)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")

	var result map[string]interface{}
	fullURL := server.URL + "/test-endpoint"
	err := client.fetchAndCache(fullURL, &result)

	assert.NoError(t, err)
	assert.Equal(t, "test", result["name"])

	cachedData, exists := cache.Get(fullURL)
	assert.True(t, exists)
	assert.Contains(t, string(cachedData), "test")
}

func TestFetchAndCache_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")

	var result map[string]interface{}
	fullURL := server.URL + "/test-endpoint"
	err := client.fetchAndCache(fullURL, &result)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "response failed with status code: 500")
}

func TestExploreLocation_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/location-area/test-area", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
            "pokemon_encounters": [
                {
                    "pokemon": {
                        "name": "pikachu",
                        "url": "https://pokeapi.co/api/v2/pokemon/25/"
                    }
                }
            ]
        }`)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")

	result, err := client.ExploreLocation("test-area")

	assert.NoError(t, err)
	assert.Len(t, result.PokemonEncounters, 1)
	assert.Equal(t, "pikachu", result.PokemonEncounters[0].Pokemon.Name)
}

func TestExploreLocation_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")

	_, err := client.ExploreLocation("nonexistent")
	assert.Error(t, err)
}

func TestCatchPokemon_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"name": "pikachu",
			"base_experience": 112,
			"height": 4,
			"weight": 60,
			"stats": [{"base_stat": 35}],
			"types": [{"type": {"name": "electric"}}]
		}`)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")

	caught, err := client.CatchPokemon("pikachu")

	assert.NoError(t, err)
	assert.IsType(t, true, caught)
}

func TestCatchPokemon_PokemonNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")

	_, err := client.CatchPokemon("fakemon")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "response failed with status code")
}

func TestGetLocationAreas_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/location-area", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"count": 42,
			"next": "https://example.com/next",
			"previous": null,
			"results": [{"name": "some-area", "url": "https://example.com/area/1"}]
		}`)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")

	result, err := client.GetLocationAreas(server.URL + "/location-area")
	assert.NoError(t, err)
	assert.Equal(t, 42, result.Count)
	assert.Equal(t, "https://example.com/next", *result.Next)
	assert.Nil(t, result.Previous)
	assert.Len(t, result.Results, 1)
	assert.Equal(t, "some-area", result.Results[0].Name)
	assert.Equal(t, "https://example.com/area/1", result.Results[0].URL)

}

func TestGetLocationAreas_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")

	result, err := client.GetLocationAreas(server.URL + "/location-area")

	assert.Error(t, err)
	assert.Equal(t, LocationAreasResponse{}, result)
}

func TestGetLocationAreas_InvalidJSON(t *testing.T) {
	cases := []struct {
		name string
		json string
	}{
		{"missing brace", `{"count": 42`},
		{"random text", `this is not json`},
		{"missing quotes", `{count: 42}`},
		{"missing colon", `{count 42}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("content-type", "application/json")
				fmt.Fprintln(w, c.json)
			}))
			defer server.Close()

			cache := pokecache.NewCache(5 * time.Minute)
			client := NewClientWithBaseURL(&cache, server.URL+"/")

			result, err := client.GetLocationAreas(server.URL + "/location-area")

			assert.Error(t, err)
			assert.Equal(t, LocationAreasResponse{}, result)

		})
	}
}

func TestInspectPokemon_PokemonNotInPokedex(t *testing.T) {
	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClient(&cache)

	result, err := client.InspectPokemon("pikachu")
	assert.NoError(t, err)
	assert.False(t, result)
}

func TestInspectPokemon_PokemonInPokedex(t *testing.T) {
	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClient(&cache)

	client.Pokedex["pikachu"] = Pokemon{
		Name:           "pikachu",
		BaseExperience: 112,
		Height:         4,
		Weight:         60,
	}

	result, err := client.InspectPokemon("pikachu")
	assert.NoError(t, err)
	assert.True(t, result)
}

func TestCatchPokemon_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")

	result, err := client.CatchPokemon("pikachu")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Equal(t, false, result)

}

func TestCatchPokemon_InvalidJSON(t *testing.T) {
	cases := []struct {
		name string
		json string
	}{
		{"missing brace", `{"name": "pikachu", "base_experience": 112`},
		{"random text", `this is not pokemon json`},
		{"missing quotes", `{name: "pikachu"}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("content-type", "application/json")
				fmt.Fprintln(w, c.json)
			}))
			defer server.Close()

			cache := pokecache.NewCache(5 * time.Minute)
			client := NewClientWithBaseURL(&cache, server.URL+"/")

			result, err := client.CatchPokemon("pikachu")

			assert.Error(t, err)
			assert.Equal(t, false, result)

		})
	}
}

func TestCatchPokemon_SuccessfulCatch_ChecksPokedex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		fmt.Fprintln(w, `{
			"name": "pikachu",
			"base_experience": 0,
			"height": 4,
			"weight": 60,
			"stats": [{"base_stat": 35}],
			"types": [{"type": {"name": "electric"}}]
		}`)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")

	var caught bool
	var err error
	for i := 0; i < 20; i++ {
		caught, err = client.CatchPokemon("pikachu")
		if caught {
			break
		}
	}

	assert.True(t, caught)
	assert.NoError(t, err)

	result, err := client.InspectPokemon("pikachu")
	assert.NoError(t, err)
	assert.True(t, result)

	_, inPokedex := client.Pokedex["pikachu"]
	assert.True(t, inPokedex)
}

func TestCatchPokemon_FailedCatch_InspectShowsNotCaught(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"name": "pikachu",
			"base_experience": 10000,
			"height": 4,
			"weight": 60,
			"stats": [{"base_stat": 35}],
			"types": [{"type": {"name": "electric"}}]
			}`)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")

	caught, err := client.CatchPokemon("pikachu")
	assert.False(t, caught)
	assert.NoError(t, err)

	expectedURL := server.URL + "/pokemon/pikachu"
	_, exists := cache.Get(expectedURL)
	assert.True(t, exists)

	_, inPokedex := client.Pokedex["pikachu"]
	assert.False(t, inPokedex)

	result, err := client.InspectPokemon("pikachu")
	assert.NoError(t, err)
	assert.False(t, result)
}

func TestCatchPokemon_FailedCatch_CachesDataButNotInPokedex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
			"name": "pikachu",
			"base_experience": 10000,
			"height": 4,
			"weight": 60,
			"stats": [{"base_stat": 35}],
			"types": [{"type": {"name": "electric"}}]
			}`)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")
	caught, err := client.CatchPokemon("pikachu")

	assert.False(t, caught)
	assert.NoError(t, err)

	_, inPokedex := client.Pokedex["pikachu"]
	assert.False(t, inPokedex)

	expectedURL := server.URL + "/pokemon/pikachu"
	_, exists := cache.Get(expectedURL)
	assert.True(t, exists)
}

func TestCatchPokemon_SecondAttempt_UsesCachedData(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{
            "name": "pikachu",
            "base_experience": 1,
            "height": 4,
            "weight": 60,
            "stats": [{"base_stat": 35}],
            "types": [{"type": {"name": "electric"}}]
        }`)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")

	_, err1 := client.CatchPokemon("pikachu")
	assert.NoError(t, err1)
	assert.Equal(t, 1, callCount)

	_, err2 := client.CatchPokemon("pikachu")
	assert.NoError(t, err2)
	assert.Equal(t, 1, callCount)

}

func TestFetchAndCache_InvalidJSON(t *testing.T) {
	cases := []struct {
		name string
		json string
	}{
		{"missing brace", `{"count": 42`},
		{"random text", `this is not json`},
		{"missing quotes", `{count: 42}`},
		{"missing colon", `{count 42}`},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("content-type", "application/json")
				fmt.Fprintln(w, c.json)
			}))
			defer server.Close()

			cache := pokecache.NewCache(5 * time.Minute)
			client := NewClientWithBaseURL(&cache, server.URL+"/")

			var result map[string]interface{}
			err := client.fetchAndCache(server.URL+"/", &result)

			assert.Error(t, err)
			assert.Empty(t, result)
		})
	}
}

func TestClient_CacheIntegration(t *testing.T) {
	callCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("content-type", "application/json")
		fmt.Fprintln(w, `{
			"count": 42,
			"next": "https://example.com/next",
			"previous": null,
			"results": [
			{"name": "some-area", "url": "https://example.com/area/1"}]
		}`)
	}))
	defer server.Close()

	cache := pokecache.NewCache(5 * time.Minute)
	client := NewClientWithBaseURL(&cache, server.URL+"/")

	result1, err1 := client.GetLocationAreas(server.URL + "/location-area")

	assert.Equal(t, 1, callCount)
	cachedData, exists := cache.Get(server.URL + "/location-area")
	assert.True(t, exists)
	assert.NotEmpty(t, cachedData)
	assert.Contains(t, string(cachedData), `"count": 42`)

	result2, err2 := client.GetLocationAreas(server.URL + "/location-area")

	assert.Equal(t, 1, callCount)
	assert.Equal(t, result1, result2)

	assert.NoError(t, err1)
	assert.NoError(t, err2)

}

func TestCatchPokemon_CatchProbabilityCalculation(t *testing.T) {
	cases := []struct {
		name           string
		baseExperience int
		expectedChance int
	}{
		{"normal case", 100, 50},
		{"minimum clamp", 600, 5},
		{"low experience", 50, 55},
		{"zero experience", 0, 60},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			actual := calculateCatchChance(c.baseExperience)
			assert.Equal(t, c.expectedChance, actual)
		})
	}
}

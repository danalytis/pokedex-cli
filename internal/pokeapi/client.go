package pokeapi

import (
	"encoding/json"
	"fmt"
	"github.com/danalytis/pokedexcli/internal/pokecache"
	"io"
	"math/rand"
	"net/http"
)

type Client struct {
	PokeapiBaseURL string
	Cache          *pokecache.Cache
	Pokedex        map[string]Pokemon
}
type Stat struct {
	BaseStat int `json:"base_stat"`
}

type Type struct {
	Type struct {
		Name string `json:"name"`
	} `json:"type"`
}

type Pokemon struct {
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Weight         int    `json:"weight"`
	Stats          []Stat `json:"stats"`
	Types          []Type `json:"types"`
}

type LocationPokemon struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type LocationPokemonEncounter struct {
	Pokemon LocationPokemon `json:"pokemon"`
}
type ExploreLocationResponse struct {
	PokemonEncounters []LocationPokemonEncounter `json:"pokemon_encounters"`
}
type LocationAreasResponse struct {
	Count    int            `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []LocationArea `json:"results"`
}

type LocationArea struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func NewClient(cache *pokecache.Cache) *Client {
	return &Client{
		PokeapiBaseURL: "https://pokeapi.co/api/v2/",
		Cache:          cache,
		Pokedex:        make(map[string]Pokemon),
	}
}

func (c *Client) fetchAndCache(url string, target interface{}) error {
	result, ok := c.Cache.Get(url)
	if !ok {
		// Cache miss - make HTTP request
		res, err := http.Get(url)
		if err != nil {
			return fmt.Errorf("error making request: %w", err)
		}
		defer res.Body.Close()

		if res.StatusCode > 299 {
			return fmt.Errorf("response failed with status code: %d", res.StatusCode)
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}

		c.Cache.Add(url, body)
		return json.Unmarshal(body, target)
	} else {
		// Cache hit - unmarshal from cache
		return json.Unmarshal(result, target)
	}
}

func calculateCatchChance(baseExperience int) int {
	maxCatchChance := 60
	catchChance := maxCatchChance - (baseExperience / 10)
	if catchChance < 5 {
		catchChance = 5
	}
	return catchChance
}

func (c *Client) CatchPokemon(name string) (bool, error) {
	url := c.PokeapiBaseURL + "pokemon/" + name

	res, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == 404 {
		return false, fmt.Errorf("%s is not a pokemon", name)
	} else if res.StatusCode > 299 {
		return false, fmt.Errorf("response failed with status code: %d", res.StatusCode)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	var pokemon Pokemon
	err = json.Unmarshal(body, &pokemon)
	if err != nil {
		return false, err
	}

	catchChance := calculateCatchChance(pokemon.BaseExperience)
	randomRoll := rand.Intn(100)

	if randomRoll < catchChance {
		c.Cache.Add(name, body)
		c.Pokedex[name] = pokemon
	}
	return randomRoll < catchChance, nil
}

func (c *Client) InspectPokemon(name string) (bool, error) {

	statNames := []string{"hp", "attack", "defense", "special-attack", "special-defense", "speed"}

	var pokemon Pokemon

	result, ok := c.Cache.Get(name)

	if !ok {
		fmt.Println("You have not caught this pokemon yet..")
		fmt.Printf("Name: %s\nHeight: ??\nWeight: ??\n", name)

		fmt.Println("Stats:")
		for i, _ := range statNames {
			fmt.Printf(" - %s: ??\n", statNames[i])
		}
		fmt.Println("Types:")
		typeName := "??"
		fmt.Printf(" - %s\n", typeName)
		return false, nil
	}

	err := json.Unmarshal(result, &pokemon)
	if err != nil {
		return false, fmt.Errorf("error unmarshaling cache data %w", err)
	}
	fmt.Printf("Name: %s\nHeight: %d\nWeight: %d\n",
		pokemon.Name,
		pokemon.Height,
		pokemon.Weight)

	fmt.Println("Stats:")
	for i, val := range pokemon.Stats {
		fmt.Printf(" - %s: %d\n", statNames[i], val.BaseStat)
	}
	fmt.Println("Types:")
	for _, typeName := range pokemon.Types {
		fmt.Printf(" - %s\n", typeName.Type.Name)
	}

	return true, nil
}

func (c *Client) ExploreLocation(name string) (ExploreLocationResponse, error) {
	url := c.PokeapiBaseURL + "location-area/" + name
	var exploreLocationResp ExploreLocationResponse

	err := c.fetchAndCache(url, &exploreLocationResp)
	if err != nil {
		return ExploreLocationResponse{}, err
	}

	return exploreLocationResp, nil
}

func (c *Client) GetLocationAreas(url string) (LocationAreasResponse, error) {
	var locationAreasResp LocationAreasResponse

	err := c.fetchAndCache(url, &locationAreasResp)
	if err != nil {
		return LocationAreasResponse{}, err
	}

	return locationAreasResp, nil
}

func NewClientWithBaseURL(cache *pokecache.Cache, baseURL string) *Client {
	return &Client{
		PokeapiBaseURL: baseURL,
		Cache:          cache,
		Pokedex:        make(map[string]Pokemon),
	}
}

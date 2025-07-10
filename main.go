package main

import (
	"bufio"
	"fmt"
	"github.com/danalytis/pokedexcli/internal/pokeapi"
	"github.com/danalytis/pokedexcli/internal/pokecache"
	"os"
	"strings"
	"time"
)

type config struct {
	Client   *pokeapi.Client
	Next     *string
	Previous *string
}
type cliCommand struct {
	name        string
	description string
	callback    func(*config, []string) error
}

func printHelp() {
	fmt.Println("Welcome to the Pokedex!")
	for _, cmd := range cliCommands {
		fmt.Printf("- %-10s %s\n", cmd.name, cmd.description)
	}
}

func commandHelp(cfg *config, args []string) error {
	fmt.Println("Type 'help' to see this message again.")
	return nil
}

func commandPokedex(cfg *config, args []string) error {
	if len(cfg.Client.Pokedex) == 0 {
		fmt.Println("Your Pokedex is empty.")
		return nil
	}
	fmt.Println("Your Pokedex:")
	for name := range cfg.Client.Pokedex {
		fmt.Printf(" - %s\n", name)
	}
	return nil
}

func commandMap(cfg *config, args []string) error {
	url := "https://pokeapi.co/api/v2/location-area"
	if cfg.Next != nil {
		url = *cfg.Next
	}
	locationAreasResp, err := cfg.Client.GetLocationAreas(url)
	if err != nil {
		return err
	}
	cfg.Next = locationAreasResp.Next
	cfg.Previous = locationAreasResp.Previous

	for _, area := range locationAreasResp.Results {
		fmt.Println(area.Name)
	}
	return nil
}

func commandMapb(cfg *config, args []string) error {
	if cfg.Previous == nil {
		fmt.Println("you're on the first page")
		return nil
	}

	locationAreasResp, err := cfg.Client.GetLocationAreas(*cfg.Previous)
	if err != nil {
		return err
	}

	cfg.Next = locationAreasResp.Next
	cfg.Previous = locationAreasResp.Previous

	for _, area := range locationAreasResp.Results {
		fmt.Println(area.Name)
	}
	return nil
}

func commandExplore(cfg *config, name []string) error {
	if len(name) == 0 {
		fmt.Println("usage: explore <location-name>")
		return nil
	}
	results, err := cfg.Client.ExploreLocation(name[0])
	if err != nil {
		return err
	}

	fmt.Println("Found Pokemon:")
	for _, encounter := range results.PokemonEncounters {
		fmt.Printf("- %s\n", encounter.Pokemon.Name)
	}
	return nil
}

func commandCatch(cfg *config, name []string) error {
	if len(name) == 0 {
		fmt.Println("usage: catch <pokemon-name>")
		return nil
	}

	results, err := cfg.Client.CatchPokemon(name[0])
	if err != nil {
		return err
	}

	fmt.Printf("Throwing a Pokeball at %s...\n", name[0])
	if results {
		fmt.Printf("%s was caught!\n", name[0])
	} else {
		fmt.Printf("%s escaped!\n", name[0])
	}

	return nil
}

func commandInspect(cfg *config, name []string) error {
	if len(name) == 0 {
		fmt.Println("usage: inspect <pokemon-name>")
		return nil
	}

	_, err := cfg.Client.InspectPokemon(name[0])
	if err != nil {
		return err
	}
	return nil
}

func commandExit(cfg *config, args []string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func cleanInput(text string) []string {

	text = strings.TrimSpace(text)
	text = strings.ToLower(text)

	words := strings.Fields(text)

	return words
}

var cliCommands = map[string]cliCommand{
	"exit": {
		name:        "exit",
		description: "Exit the Pokedex",
		callback:    commandExit,
	},
	"help": {
		name:        "help",
		description: "Displays a help message",
		callback:    commandHelp,
	},
	"map": {
		name:        "map",
		description: "Display nearby locations",
		callback:    commandMap,
	},
	"mapb": {
		name:        "mapb",
		description: "Display previous nearby locations",
		callback:    commandMapb,
	},
	"explore": {
		name:        "explore",
		description: "List all pokemon found in a location",
		callback:    commandExplore,
	},
	"catch": {
		name:        "catch",
		description: "Catch a pokemon",
		callback:    commandCatch,
	},
	"inspect": {
		name:        "inspect",
		description: "Prints the stats of a pokemon",
		callback:    commandInspect,
	},
	"pokedex": {
		name:        "pokedex",
		description: "Lists all pokemon in your pokedex",
		callback:    commandPokedex,
	},
}

func main() {

	cache := pokecache.NewCache(5 * time.Second)
	client := pokeapi.NewClient(&cache)

	cfg := &config{
		Client: client,
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Pokedex > ")
	for scanner.Scan() {
		command := scanner.Text()
		cleanedCommand := cleanInput(command)

		if len(cleanedCommand) == 0 {
			fmt.Print("Pokedex > ")
			continue
		}

		cmd, ok := cliCommands[cleanedCommand[0]]
		if !ok {
			fmt.Println("Unknown command")
		} else {
			if cleanedCommand[0] == "help" {
				printHelp()
			}
			err := cmd.callback(cfg, cleanedCommand[1:])
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error: ", err)
			}
		}
		fmt.Print("Pokedex > ")
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

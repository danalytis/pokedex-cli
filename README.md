# Pokedex CLI

A command-line interface for exploring the Pokemon world using the PokeAPI.

[![Boot.dev](https://img.shields.io/badge/Boot.dev-Guided%20Project-4f46e5)](https://boot.dev)

## Overview

This CLI application allows users to explore Pokemon locations, catch Pokemon, and manage a personal Pokedex collection. Built as part of the Boot.dev backend development curriculum.

## Quick Start

```bash
git clone https://github.com/yourusername/pokedexcli.git
cd pokedexcli
go build
./pokedexcli
```

## Commands

- `help` - Show available commands
- `map` - List nearby locations
- `mapb` - Show previous locations
- `explore <location>` - Find Pokemon in a location
- `catch <pokemon>` - Attempt to catch a Pokemon
- `inspect <pokemon>` - View caught Pokemon details
- `pokedex` - List your collection
- `exit` - Quit the application

## Features

- Location-based Pokemon discovery
- Probability-based catching mechanics
- HTTP response caching
- Personal Pokemon collection

## Testing

```bash
go test ./...
```

## Built With

- Go
- [PokeAPI](https://pokeapi.co/)
- HTTP client with caching layer

package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCleanInput_Uppercase(t *testing.T) {
	result := cleanInput("EXPLORE PALLET-TOWN")
	expected := []string{"explore", "pallet-town"}
	assert.Equal(t, expected, result)
}

func TestCleanInput_SingleWord(t *testing.T) {
	result := cleanInput("help")
	expected := []string{"help"}
	assert.Equal(t, expected, result)
}

func TestCleanInput_EmptyString(t *testing.T) {
	result := cleanInput("")
	assert.Equal(t, 0, len(result))
}

func TestCleanInput_ExtraWhitespaces(t *testing.T) {
	result := cleanInput("   explore   pallet-town  ")
	expected := []string{"explore", "pallet-town"}
	assert.Equal(t, expected, result)
}

func TestCommandHelp(t *testing.T) {
	cfg := &config{}
	err := commandHelp(cfg, []string{})
	assert.NoError(t, err)
}

package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bladetusk Boar — "Intimidate (This creature can't be blocked except by
// artifact creatures and/or creatures that share a color with it.)"
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f640a102-7ee8-4605-9e15-09938fb229b2",
		Name:            "Bladetusk Boar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
	})
}

package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreg Mangler — Creature — Plant Zombie {1}{B}{G}, 3/3 (Double Masters 2022):
//
//	"Haste
//	 Scavenge {3}{B}{G} ({3}{B}{G}, Exile this card from your graveyard: Put a
//	 number of +1/+1 counters equal to this card's power on target creature.
//	 Scavenge only as a sorcery.)"
//
// Haste is a printed keyword; scavenge is the engine's (scavenge.go), with
// the counters read off the card the cost exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "71358110-4e28-4794-befa-ea6e27e720ef",
		Name:            "Dreg Mangler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated:       []effects.ActivatedAbility{effects.Scavenge("{3}{B}{G}")},
	})
}

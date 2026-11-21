package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vermin Gorger — Creature — Vampire {1}{B}, 2/2 (Modern Horizons 2,
// #5501):
//
//	"{T}, Sacrifice another creature: Each opponent loses 2 life and you
//	 gain 2 life."
//
// "Another" is object identity (SacrificeAnotherN, #1738); the {T} waits
// out summoning sickness (CR 302.6). The drain is Palace Siege's Dragons
// body: you gain 2 whatever the opponents lost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b81596b-9487-4be1-a417-6913fa72299d",
		Name:         "Vermin Gorger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice another creature: Each opponent loses 2 life and you gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect:  effects.PalaceSiegeDrain,
		}},
	})
}

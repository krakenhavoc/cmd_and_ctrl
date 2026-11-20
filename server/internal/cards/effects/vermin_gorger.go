package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7b81596b-9487-4be1-a417-6913fa72299d",
		Name:         "Vermin Gorger",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice another creature: Each opponent loses 2 life and you gain 2 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(TapCost(), SacrificeAnotherN(1, "another creature", Creature())),
			Effect:  palaceSiegeDrain,
		}},
	})
}

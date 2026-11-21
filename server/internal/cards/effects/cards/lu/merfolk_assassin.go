package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merfolk Assassin — Creature — Merfolk Assassin {U}{U}, 1/2
// (The Dark, #5638):
//
//	"{T}: Destroy target creature with islandwalk."
//
// Royal Assassin's row with a keyword clause in place of "tapped".
// HasKeyword reads the creature's current abilities, so islandwalk
// granted by another effect makes a creature a legal target, and a
// creature that loses islandwalk in response is no longer one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7edd536c-c043-4189-aa57-83d22eed3103",
		Name:         "Merfolk Assassin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Destroy target creature with islandwalk.",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature with islandwalk", effects.HasKeyword("islandwalk")),
			Effect:  effects.DestroyFirstLegalTarget,
		}},
	})
}

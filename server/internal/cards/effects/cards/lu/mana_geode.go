package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mana Geode — Artifact {3}:
//
//	"When this artifact enters, scry 1.
//	 {T}: Add one mana of any color."
//
// A Commander's Sphere with a scry on entry. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0844f4e6-2b98-4d09-a5c5-92f0a3b6a517",
		Name:         "Mana Geode",
		Purpose:      game.Purpose{Scry: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mana Geode — scry 1", effects.Do(effects.Scry{N: 1})),
		},
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Tap: true}, "{T}: Add one mana of any color"),
		},
	})
}

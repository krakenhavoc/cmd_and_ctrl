package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rosheen Meanderer — Legendary Creature — Giant Shaman {3}{R/G}, 4/4
// (Shadowmoor, #5591):
//
//	"{T}: Add {C}{C}{C}{C}. Spend this mana only on costs that contain
//	 {X}."
//
// A tap mana ability whose mana carries Elementalist's Palette's
// restriction (game.ManaRestrictXCost, CR 106.6): it may pay the whole of a
// spell's cost when that spell's mana cost has {X}, or of an activated
// ability's cost with {X} in it, and nothing else.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "892d318e-34bd-4fca-a2a1-0c3534c68068",
		Name:         "Rosheen Meanderer",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{C}{C}{C}{C}",
			Restrictions: []string{game.ManaRestrictXCost},
			Label:        "{T}: Add {C}{C}{C}{C}. Spend this mana only on costs that contain {X}.",
		}},
	})
}

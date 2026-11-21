package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Groundchuck & Dirtbag — {4}{G}{G} Legendary Creature — Ox Mole Mutant 8/8
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Trample
//	 Whenever you tap a land for mana, add {G}."
//
// Mirari's Wake's one-sided mana trigger with a fixed {G}: a triggered mana
// ability (CR 605.1b) that resolves at once, off the stack, whenever its
// controller taps a land they control for mana. An opponent's land adds
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "da344e46-6bdf-4032-8590-bac850d5713d",
		Name:            "Groundchuck & Dirtbag",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		ManaTriggers: []game.ManaTrigger{
			effects.WheneverYouTapALandForMana("Groundchuck & Dirtbag — add {G}", effects.AddsFixedMana("{G}")),
		},
	})
}

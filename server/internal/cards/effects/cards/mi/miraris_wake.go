package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirari's Wake — Enchantment {3}{G}{W}:
//
//	"Creatures you control get +1/+1.
//	 Whenever you tap a land for mana, add one mana of any type that
//	 land produced."
//
// The anthem is Glorious Anthem's layer-7c static. The second
// ability is a TRIGGERED MANA ability (CR 605.1b) — the shape Zendikar
// Resurgent's file names as Mirari's Wake's — which fires on the tap,
// adds mana and never uses the stack (CR 605.4a). "Of any type that
// land produced" includes colourless.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "852657c0-18a4-4b28-b9ae-7728acdb5044",
		Name:         "Mirari's Wake",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.TribalAnthem(effects.TribeFilter{YoursOnly: true}, 1, 1)},
		ManaTriggers: []game.ManaTrigger{
			effects.WheneverYouTapALandForMana(
				"Mirari's Wake — add one mana of any type that land produced",
				effects.AddsOneManaOfAnyTypeProduced()),
		},
	})
}

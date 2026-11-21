package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gibbering Barricade — Creature — Nightmare Wall {2}{B}, 2/4 (Dominaria
// United):
//
//	"Defender
//	 {2}{B}, Sacrifice a creature: You gain 1 life and draw a card."
//
// A sacrifice outlet (Viscera Seer's cost; it may sacrifice itself). The
// life and the card are one effect, in that order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c6106021-08ae-49c8-b32c-fdc7397522e7",
		Name:            "Gibbering Barricade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice a creature: You gain 1 life and draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{2}{B}"), effects.SacrificeACreature()),
			Effect:  effects.Do(effects.GainLife{Amount: 1}, effects.DrawCards{N: 1}),
		}},
	})
}

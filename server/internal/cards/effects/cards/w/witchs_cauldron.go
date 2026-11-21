package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Witch's Cauldron — Artifact {B} (Jumpstart):
//
//	"{1}{B}, {T}, Sacrifice a creature: You gain 1 life and draw a card."
//
// Vampiric Rites with a tap in the cost, so once per untap. The
// sacrificed creature is chosen as the cost is paid and the gain comes
// before the draw (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb65ee08-bfdb-4c18-8967-8bcf31fcabfa",
		Name:         "Witch's Cauldron",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}, {T}, Sacrifice a creature: You gain 1 life and draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, Draws: 1},
			Cost:    effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost(), effects.SacrificeACreature()),
			Effect:  effects.Do(effects.GainLife{Amount: 1}, effects.DrawCards{N: 1}),
		}},
	})
}

package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Krosan Wayfarer — Creature — Human Druid {G}, 1/1 (Judgment, #5611):
//
//	"Sacrifice this creature: You may put a land card from your hand onto the
//	 battlefield."
//
// The sacrifice is the cost; the land is the controller's optional pick from
// their hand, put onto the battlefield untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1729317d-ccfb-4f27-84c6-20b36b2826c9",
		Name:         "Krosan Wayfarer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: You may put a land card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1, LandsUntapped: 1},
			Cost:    effects.SacrificeThis(),
			Effect:  effects.Do(effects.MayPutALandFromHand("Krosan Wayfarer")),
		}},
	})
}

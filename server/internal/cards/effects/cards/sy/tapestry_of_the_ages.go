package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tapestry of the Ages — Artifact {4} (Dragons of Tarkir, S82 old batch 329):
//
//	"{2}, {T}: Draw a card. Activate only if you've cast a noncreature spell
//	 this turn."
//
// Seeker of Insight's activation condition, read live (CR 602.5b) from the
// controller's cast tally for the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9248f9b9-ddaa-459c-91e6-9045a65b0f07",
		Name:         "Tapestry of the Ages",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{2}, {T}: Draw a card. Activate only if you've cast a noncreature spell this turn.",
			Cost:      effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool { return g.CastTallyFor(controller).Noncreature > 0 },
			Purpose:   game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Effect:    effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}

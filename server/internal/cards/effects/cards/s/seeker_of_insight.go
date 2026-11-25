package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seeker of Insight — {1}{U} Creature — Human Wizard 1/3 (Amonkhet, S82 old
// batch 245):
//
//	"{T}: Draw a card, then discard a card. Activate only if you've cast a
//	 noncreature spell this turn."
//
// The restriction is an activation condition read live (CR 602.5b), from the
// controller's cast tally for the turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a0b2d446-fe90-40f8-8674-7421c425b489",
		Name:         "Seeker of Insight",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{T}: Draw a card, then discard a card. Activate only if you've cast a noncreature spell this turn.",
			Cost:      effects.TapCost(),
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool { return g.CastTallyFor(controller).Noncreature > 0 },
			Purpose:   game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, 1, 1, "")
			},
		}},
	})
}

package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Corpse Churn — Instant {1}{B} (Ravnica: Clue Edition, #5929):
//
//	"Mill three cards, then you may return a creature card from your
//	 graveyard to your hand."
//
// Dig Up the Body's shape: the mill first, so a milled creature can be
// returned; the choice is optional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e2aeac0b-af74-4c47-9ac1-51fbf869c95f",
		Name:         "Corpse Churn",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.MillCards{N: 3}.Apply(ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.ReturnChosenFromGraveyard{
					Player:   ctx.Controller(),
					Question: "Corpse Churn — you may return a creature card from your graveyard to your hand",
					Match:    game.Card.IsCreature,
					Max:      1,
					ToHand:   true,
				}.Apply(ctx)
			}),
		),
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e2aeac0b-af74-4c47-9ac1-51fbf869c95f",
		Name:         "Corpse Churn",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return MillCards{N: 3}.Apply(ctx) }),
			Instruction(func(ctx *Context) error {
				return ReturnChosenFromGraveyard{
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

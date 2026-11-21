package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sokka's Haiku — Instant — Lesson {3}{U}{U} (Avatar: The Last
// Airbender, #4459):
//
//	"Counter target spell.
//	 Draw a card, then mill three cards.
//	 Untap target land."
//
// Two target clauses (Clauses, CR 601.2c): the spell and the land. Each
// is re-checked as it resolves (CR 608.2b): a spell that left the stack
// is not countered, and a land that left is not untapped, but the rest
// still happens while either target is legal. Only if both are illegal
// is the Haiku itself not resolved at all (CR 608.2b). The draw lands
// before the mill.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87a97c72-143d-48a8-a1c5-3c44aa8e5a86",
		Name:         "Sokka's Haiku",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetSpell("target spell"),
			effects.TargetPermanent("target land", effects.Land()),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
					if err := (effects.CounterTarget{StackID: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return effects.DrawCards{N: 1}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error { return effects.MillCards{N: 3}.Apply(ctx) }),
			// The untap is its own instruction: it reads the land target,
			// never the mill's outcome.
			effects.Instruction(func(ctx *effects.Context) error {
				if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
					return effects.UntapTarget{Target: t.ID}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}

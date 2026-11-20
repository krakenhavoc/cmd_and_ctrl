package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Read the Tides — Sorcery {5}{U} (Core Set 2021, #5511):
//
//	"Choose one —
//	 • Draw three cards.
//	 • Return up to two target creatures to their owners' hands."
//
// Two ModeDoing bullets. The bounce is Calamitous Tide's "up to two"
// clause, returned together (one simultaneous move); a creature that left
// in response is skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e5c4ed40-5069-4606-81f2-59dcfa1345bd",
		Name:         "Read the Tides",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Draw three cards.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return DrawCards{N: 3}.Apply(ctx)
				}),
			ModeDoing("Return up to two target creatures to their owners' hands.",
				TargetCreature("up to two target creatures").WithCount(0, 2),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					var ids []uuid.UUID
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
							ids = append(ids, t.ID)
						}
					}
					ctx.Game.BounceCardsToHandForEffect(ids)
					return nil
				}),
		),
	})
}

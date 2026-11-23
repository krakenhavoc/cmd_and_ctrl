package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "e5c4ed40-5069-4606-81f2-59dcfa1345bd",
		Name:         "Read the Tides",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Draw three cards.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DrawCards{N: 3}.Apply(ctx)
				}),
			effects.ModeDoing("Return up to two target creatures to their owners' hands.",
				effects.TargetCreature("up to two target creatures").WithCount(0, 2),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
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

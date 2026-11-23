package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Careful Consideration — Instant {2}{U}{U} (Time Spiral, #5596):
//
//	"Target player draws four cards, then discards three cards. If you
//	 cast this spell during your main phase, instead that player draws
//	 four cards, then discards two cards."
//
// drawThenDiscard on the targeted player: the discard prompt opens once
// all four draws are done, for three cards, or two when the spell was
// cast in its controller's main phase (tspB2MainPhaseAmount).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b34b361-7b0e-4464-874c-ceb501bc5ecc",
		Name:         "Careful Consideration",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 4, Discards: 3}),
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			return effects.DrawThenDiscard(ctx.Game, t.ID, item.SourceCardID, 4, effects.TspB2MainPhaseAmount(ctx, 3, 2), "")
		},
	})
}

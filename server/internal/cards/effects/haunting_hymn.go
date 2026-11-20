package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haunting Hymn — Instant {4}{B}{B} (Time Spiral, #5596):
//
//	"Target player discards two cards. If you cast this spell during your
//	 main phase, that player discards four cards instead."
//
// The targeted player picks the discards (CR 701.9a): two, or four when
// the spell was cast in its controller's main phase (tspB2MainPhaseAmount,
// read as it resolves).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9a0977a-6364-4358-bab1-112898a1b942",
		Name:         "Haunting Hymn",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(game.TargetPurpose{Slot: 0, Discards: 2}),
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetPlayer {
				return nil
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player: t.ID, Source: item.SourceCardID, N: tspB2MainPhaseAmount(ctx, 2, 4),
			})
			return nil
		},
	})
}

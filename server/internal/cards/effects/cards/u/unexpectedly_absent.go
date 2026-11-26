package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unexpectedly Absent — Instant {X}{W}{W} (Commander 2020, S82 OLD, #5513):
//
//	"Put target nonland permanent into its owner's library just beneath the
//	 top X cards of that library."
//
// Underdark Rift's tuck with X from the cast in place of a d10: "just
// beneath the top X cards" is X + 1 cards down (TuckOptions.Depth), and a
// library of fewer than X cards takes the permanent on the bottom (CR 401.4).
// A target that has left, or stopped being a nonland permanent, is not tucked
// (CR 608.2b); it goes to its OWNER's library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e8d78a83-c932-4b55-8f75-7094c672c3a9",
		Name:         "Unexpectedly Absent",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Not(effects.Land())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return ctx.Game.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{t.ID}, game.TuckOptions{Depth: ctx.X() + 1}, game.KeyedTail{})
			}
			return nil
		},
	})
}

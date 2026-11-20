package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mind Shatter — Sorcery {X}{B}{B} (Morningtide, #5592):
//
//	"Target player discards X cards at random."
//
// Mind Twist at a different price: the random discard (CR 701.9b) scaled
// by the announced X. X = 0 discards nothing; a hand smaller than X is
// emptied.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "26809033-a823-4251-887e-0410a545f993",
		Name:         "Mind Shatter",
		XMatters:     true,
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    mindShatterDiscard,
	})
}

// mindShatterDiscard is "target player discards X cards at random".
func mindShatterDiscard(_ *game.StackItem, ctx *Context) error {
	if p := TargetedPlayer(ctx); p != uuid.Nil {
		return DiscardCards{Player: p, N: ctx.X()}.Apply(ctx)
	}
	return nil
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// skip_turns.go — the card side of CR 614.10 "skip your next turn" (ADR
// 0188 Decision 1, #3559, tracker #753).
//
// "You skip your next turn" is SkipYourNextTurns, and "target player
// skips their next X turns" is TargetPlayerSkipsNextTurns. Neither moves
// the turn cursor. Each writes a skip debt (game.SkipNextTurnsForEffect)
// that the engine spends at the rotation seam, as the player's next turn
// would begin: the next turn they would begin, normal or extra (the
// Magosi ruling), and never the turn already in progress (CR 614.10).
// Two skips are two turns (CR 614.10a).

// SkipYourNextTurns is "[you] skip your next turn" (N turns, one when
// left zero). "You" is the resolving item's controller, which on an
// any-player activated ability (Lethal Vapors) is whoever activated it
// (CR 109.5).
type SkipYourNextTurns struct {
	// N is how many turns; zero means one. Eater of Days says two.
	N int
}

// Apply implements Primitive. It writes a record and never fails: a
// player who has left the game skips nothing (CR 800.4a).
func (x SkipYourNextTurns) Apply(ctx *Context) error {
	skipNextTurns(ctx, ctx.Controller(), x.N)
	return nil
}

// TargetPlayerSkipsNextTurns is "target player skips their next turn"
// (N turns, one when left zero), reading the player from the card's
// player target slot. A target that left the game in response is no
// longer legal (CR 608.2b), so the effect does nothing for it.
type TargetPlayerSkipsNextTurns struct {
	// N is how many turns; zero means one.
	N int
}

// Apply implements Primitive.
func (x TargetPlayerSkipsNextTurns) Apply(ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			skipNextTurns(ctx, t.ID, x.N)
		}
	}
	return nil
}

// skipNextTurns is the one place the primitives turn "N, or one when
// zero" into a debt, labelled with the resolving item.
func skipNextTurns(ctx *Context, player uuid.UUID, n int) {
	if n == 0 {
		n = 1
	}
	label := ""
	if ctx.Item != nil {
		label = ctx.Item.Label
	}
	ctx.Game.SkipNextTurnsForEffect(player, ctx.Source(), n, label)
}

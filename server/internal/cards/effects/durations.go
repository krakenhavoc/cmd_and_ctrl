package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// durations.go — the card-facing half of the CR 611.2 duration model
// (ADR 0063, #755). `until_end_of_turn.go` has the primitives for the
// one duration that already existed; this file builds all four, for
// `ScopedEffectFor` (scoped_effect.go), which takes any of them.
//
// The builders read the game for the stamps a duration needs — the
// player's seat-turn count, the source's battlefield-entry stamp — so
// a card never has to know how a duration is represented:
//
//	DurationUntilEndOfTurn(ctx)
//	DurationUntilYourNextTurn(ctx, player)
//	DurationWhileSourceRemains(ctx, source)
//	DurationWhileYouControlSource(ctx, source, player)
//	game.IndefiniteDuration()
//
// #755 proposed three named wrappers — StaticUntilYourNextTurn,
// StaticIndefinite, StaticWhileSourceRemains. They would be three
// copies of the same four lines differing in one argument, which is
// the shape the catalog's clone-baseline guard exists to prevent. The
// closure-taking escape hatches they would have wrapped,
// `StaticForDuration` and `StaticUntilEOT`, were deleted with ADR 0041
// phase 3 tier 3a (#1497): a continuous effect from a resolution is a
// data record, and nothing accepts a closure for one any more.

// DurationUntilEndOfTurn is "until end of turn" (CR 514.2). The
// effect ends in the cleanup step of the turn it was created in —
// including one created during that turn's END step, which is not the
// end of the turn.
func DurationUntilEndOfTurn(ctx *Context) game.Duration {
	return ctx.Game.UntilEndOfTurnDuration()
}

// DurationUntilYourNextTurn is "until <player>'s next turn"
// (CR 611.2b). The effect ends as that player's next turn begins,
// before anything untaps (CR 500.1, CR 502.3) — and, if that player
// has left the game by then, at the moment their turn would have
// begun (CR 800.4m).
func DurationUntilYourNextTurn(ctx *Context, player uuid.UUID) game.Duration {
	return ctx.Game.UntilYourNextTurnDuration(player)
}

// DurationWhileSourceRemains is "for as long as ~ remains on the
// battlefield" (CR 611.2b) — Sower of Temptation.
//
// The second return is false when `source` is not on the battlefield
// right now: CR 611.2b says an effect whose condition is already
// false as it would begin never begins at all, so the caller
// registers nothing rather than an effect that would die on its first
// recompute.
//
// The condition is keyed on the source as the object it is NOW, so a
// Sower that leaves later — flickered, even — hands the creature back
// (CR 400.7). One that left and came back BEFORE its trigger resolved
// is not the Sower the trigger names, and nothing is taken at all.
func DurationWhileSourceRemains(ctx *Context, source uuid.UUID) (game.Duration, bool) {
	// #1432: a source that left and came back while the ability
	// waited is not the object "~" names, so the effect never begins.
	if ctx.isNewSourceObjectAsThis(source) {
		return game.Duration{}, false
	}
	return ctx.Game.ForAsLongAsOnBattlefieldDuration(source)
}

// DurationWhileYouControlSource is "for as long as you control ~"
// (CR 611.2b). Same "never begins" rule as its sibling, plus the
// control test — losing the source ends the effect even though the
// source is still on the battlefield.
func DurationWhileYouControlSource(ctx *Context, source, player uuid.UUID) (game.Duration, bool) {
	if ctx.isNewSourceObjectAsThis(source) { // #1432, as above
		return game.Duration{}, false
	}
	return ctx.Game.ForAsLongAsYouControlDuration(source, player)
}

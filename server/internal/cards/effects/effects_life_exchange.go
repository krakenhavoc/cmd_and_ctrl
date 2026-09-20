package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// effects_life_exchange.go — "exchange life totals" (CR 119.9), the
// clause Axis of Mortality, Magus of the Mirror and Mister Negative
// share. Tree of Perdition exchanges a life total with a TOUGHNESS,
// which is a different pair of halves and lives in its own file.
//
// There is no life-set and no life-exchange primitive, and neither is
// needed: an exchange is two "this player's life total becomes N"
// writes with BOTH totals captured before either of them runs. That
// is the whole of the rule, and b31LifeBecomes is already the
// established way to write one half — the player gains or loses the
// difference, so the CR 614 replacement window runs and every
// lifegain / life-loss payoff at the table sees the change.
//
// WHAT THIS IS NOT. One atomic event. The two halves are two
// independent ChangePlayerLifeForEffect calls, each opening its own
// replacement window, so a replacement that fires on the first half
// observes a board state a true simultaneous exchange never shows:
// one player already swapped, the other not yet. Nothing in the
// catalog replaces a life change conditionally on another player's
// total today, so the difference is not observable from any card that
// exists here — which is why it is documented rather than declared as
// a Caveat. A caveat has to be a sentence a player can act on, and
// "the two halves are sequenced" is not one until a card makes it
// visible.

// exchangeLifeTotals swaps `a`'s and `b`'s life totals. Both totals
// are read before either write, so the second half sets the value the
// first half replaced rather than the value it produced.
//
// A no-op when either player has left, when the two are the same
// player, or when the totals already match — an exchange of equal
// totals changes nothing, and emitting two zero-delta life events for
// it would hand "whenever you gain life" a change that did not
// happen.
//
// Caller holds g.mu, as every effect body does.
func exchangeLifeTotals(ctx *Context, a, b uuid.UUID) error {
	return exchangeLifeTotalsThen(ctx, a, b, nil)
}

// exchangeLifeTotalsThen is exchangeLifeTotals with a continuation
// carrying the delta that was actually APPLIED to `a` — signed the
// way the event is, negative for life lost. Mister Negative's "if you
// lost life this way, draw that many cards" reads it.
//
// It is the applied amount rather than arithmetic on the two totals
// because a replacement can change what really happened: under a
// "your life total can't change" effect the exchange moves nothing
// and the draw is nothing, which subtracting the raw totals would get
// wrong in the card's favour.
//
// `then` runs once, after BOTH halves, and is told zero on every path
// where no life moved — a continuation that is never called is a
// resolution that never finishes.
func exchangeLifeTotalsThen(ctx *Context, a, b uuid.UUID, then func(g *game.Game, appliedToA int) error) error {
	finish := func(g *game.Game, applied int) error {
		if then == nil {
			return nil
		}
		return then(g, applied)
	}
	if a == uuid.Nil || b == uuid.Nil || a == b {
		return finish(ctx.Game, 0)
	}
	pa, pb := ctx.PlayerByID(a), ctx.PlayerByID(b)
	if pa == nil || pb == nil {
		return finish(ctx.Game, 0)
	}
	lifeA, lifeB := pa.Life, pb.Life
	if lifeA == lifeB {
		return finish(ctx.Game, 0)
	}
	source := ctx.Source()
	return ctx.Game.ChangePlayerLifeThenForEffect(source, a, lifeB-lifeA, func(g *game.Game, applied int) error {
		// The second half is written from the CAPTURED total, not
		// from `applied`: "becomes lifeA" is what the exchange says,
		// whatever the first half's replacements did to the first
		// player. Re-read b's live total so the delta lands on
		// whatever it is now — a life replacement on the first half
		// can have moved it.
		if p := g.PlayerByIDForEffect(b); p != nil && p.Life != lifeA {
			if err := g.ChangePlayerLifeForEffect(source, b, lifeA-p.Life); err != nil {
				return err
			}
		}
		return finish(g, applied)
	})
}

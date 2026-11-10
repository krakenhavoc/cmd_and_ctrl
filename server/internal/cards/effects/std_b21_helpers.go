package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b21_helpers.go — shared pieces for the Standard set leftovers,
// batches 21 and 22 (#4452).

// copyYourNextInstantOrSorcery is the OnResolve of "When you next cast an
// instant or sorcery spell this turn, copy that spell. You may choose new
// targets for the copy." (Doublecast's text on an instant, Teach by
// Example): a delayed trigger that waits for the next such cast (CR 603.7b)
// and ends with the turn.
func copyYourNextInstantOrSorcery(label string) func(*game.StackItem, *Context) error {
	return func(_ *game.StackItem, ctx *Context) error {
		return WhenYouNextCast(label,
			game.CastFilter{Types: []string{"Instant", "Sorcery"}}, copyTheSpellBody).Apply(ctx)
	}
}

// tapAndStunTheTarget is "tap target creature an opponent controls and put
// a stun counter on it" as a trigger's effect (Alchemax Slayer-Bots,
// Floodpits Drowner): the first legal target is tapped, then given the
// counter.
func tapAndStunTheTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	id := ts[0].ID
	if err := (TapTarget{Target: id}).Apply(ctx); err != nil {
		return err
	}
	return (AddCounter{Target: id, Kind: game.CounterStun, N: 1}).Apply(ctx)
}

// hasAStunCounter is "creature with a stun counter on it".
func hasAStunCounter() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
		return c.Counters[game.CounterStun] > 0
	}
}

// bounceTheModesTargets is "return <this mode's targets> to their owners'
// hands" as a modal bullet's body (Atlantis Attacks, Get Out): each target
// still legal at resolution (CR 608.2b) goes to its OWNER's hand, and one that
// is no longer legal is skipped while the rest still return.
func bounceTheModesTargets(_ *game.StackItem, ctx *Context, occ int) error {
	for _, t := range ctx.ModeTargets(occ) {
		if !ctx.IsTargetLegal(t) {
			continue
		}
		if err := (BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

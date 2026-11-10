package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Break the Spell — Instant {W} (Wilds of Eldraine, #4471):
//
//	"Destroy target enchantment. If a permanent you controlled or a token was
//	 destroyed this way, draw a card."
//
// The destruction is the keyed batch form (Claim the Precious), whose tail is
// told what was actually destroyed: an indestructible enchantment is not, so
// nothing is drawn (the ruling). Whether it was "a permanent you controlled or
// a token" is read before the move, while the enchantment is still on the
// battlefield (a token ceases to exist afterwards, and the controller is the
// last-known one, CR 608.2h), and rides the tail as its N. A permanent the
// destruction sent somewhere other than a graveyard still counts (the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9d51e07-851f-4b7b-bffe-7c1efff0714c",
		Name:         "Break the Spell",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target enchantment", Enchantment()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			qualifies := 0
			if c.Controller == ctx.Controller() || c.IsToken() {
				qualifies = 1
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{id}, ctx.TailN(breakTheSpellDraw, qualifies))
		},
	})
}

// breakTheSpellDraw is "if a permanent you controlled or a token was destroyed
// this way, draw a card": told the destroyed cards, with N saying whether the
// one it was about qualified. An on-disk key: never renamed, never reused.
var breakTheSpellDraw = CardTail("card/break-the-spell-draw", func(ctx *Context, o game.TailOutcome) error {
	if o.Tail.N == 0 || len(o.IDs) == 0 {
		return nil
	}
	return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})

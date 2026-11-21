package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// First Contact — Instant {1}{G} (Star Trek, #4973):
//
//	"Mill four cards. You may put a permanent card from among them into
//	 your hand. If this spell is the first spell you've cast this game,
//	 you gain 2 life."
//
// Something Worth Saving's mill-then-take: the pick is over the cards
// that actually reached the graveyard (CR 400.7). "The first spell
// you've cast this game" is the very object stamped as its controller's
// first cast (Player.FirstSpellCast, CR 601.2i), read as the spell begins
// to resolve, so a spell cast in response does not change it; a copy is
// not cast (CR 707.10) and gains nothing. The life is gained after the
// pick, in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81229e0a-8413-40cf-8087-9d580cbb21c3",
		Name:         "First Contact",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			then := firstContactMilled
			if effects.TrekB20IsFirstSpellCastThisGame(ctx) {
				then = firstContactMilledFirstSpell
			}
			return effects.RfSpellBMillThenMayTakePermanent(ctx,
				"First Contact — put a permanent card milled this way into your hand", then)
		},
	})
}

// firstContactMilled is the mill's keyed "then" when First Contact is not
// its controller's first spell: the pick alone.
var firstContactMilled = effects.RfSpellBTakeFromMilledTail("mill-then-take/first-contact", func(*effects.Context) error { return nil })

// firstContactMilledFirstSpell is the pick, then "you gain 2 life".
var firstContactMilledFirstSpell = effects.RfSpellBTakeFromMilledTail("mill-then-take/first-contact-first-spell", func(c *effects.Context) error {
	return effects.GainLife{Player: c.Controller(), Amount: 2}.Apply(c)
})

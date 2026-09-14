package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ox of Agonas — Creature — Ox {3}{R}{R}, 4/2:
//
//	"When this creature enters, discard your hand, then draw three
//	 cards.
//	 Escape—{R}{R}, Exile eight other cards from your graveyard.
//	 This creature escapes with a +1/+1 counter on it."
//
// The one escape card whose escape cost is CHEAPER than its printed
// cost, and the reason the engine prices the two paths separately
// rather than layering a surcharge: {3}{R}{R} from hand, {R}{R} from
// the graveyard, and eight cards is what buys the difference. A deck
// that can pay it is a deck that would rather have the cards in
// exile anyway.
//
// The ETB is a wheel, and the ordering is the card: the discard
// happens first and lands in the same graveyard the escape cost was
// just paid out of, then three fresh cards arrive. Discarding an
// empty hand is a legal no-op, which is the ordinary case after an
// escape.
//
// Sandbox simplification: the discard picks nothing, because there
// is nothing to pick — "discard your hand" takes every card, so the
// engine's random-selection discard is exactly right here and the
// usual caveat about it does not apply.
func init() {
	Register(Spec{
		OracleID:         "22113051-c971-4108-953b-95356d21323a",
		Name:             "Ox of Agonas",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{EscapeWithCounters("{R}{R}", 8, 1)},
		OnETB: func(card *game.Card, ctx *Context) error {
			if _, err := discardWholeHand(ctx.Game, card.Controller); err != nil {
				return err
			}
			return DrawCards{Player: card.Controller, N: 3}.Apply(ctx)
		},
	})
}

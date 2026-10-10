package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Twinned Vision — Instant {1}{U/R} (Reality Fracture, tracker #2795):
//
//	"Draw a card. If this spell wasn't cast from your hand, draw two
//	 cards instead.
//	 Flashback—{1}{U/R}{U/R}, Discard a card."
//
// The discard is part of the flashback PRICE (CR 702.34a), so it is a
// component of the flashback offer, not an additional cost of the card:
// a cast from hand discards nothing. It is the retrace discard
// component, which names a card from the caster's hand and discards it
// as the spell goes on the stack. "Wasn't cast from your hand" reads
// CastFromZone, so a copy (never cast) draws two as well.
//
// No simplification.
func init() {
	flashback := Flashback("{1}{U/R}{U/R}")
	flashback.Label = "Flashback—{1}{U/R}{U/R}, Discard a card"
	flashback.DiscardFromHand = CardInYourHand("a card")
	flashback.PayLabel = "a card"
	Register(Spec{
		OracleID:         "fa1cc78f-26bc-4d5d-973b-7cb7d7490973",
		Name:             "Twinned Vision",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{flashback},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := 2
			if item.CastFromZone == game.ZoneHand {
				n = 1
			}
			return DrawCards{Player: ctx.Controller(), N: n}.Apply(ctx)
		},
	})
}

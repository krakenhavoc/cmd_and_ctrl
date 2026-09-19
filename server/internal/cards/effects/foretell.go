package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// foretell.go — #658: foretell (CR 702.143).
//
// "Foretell {1}{U}" is two costs and a card file should only ever
// spell out the one that varies. The {2} you pay to take the special
// action is the KEYWORD's (CR 702.143a) and is the same on every card
// that prints foretell; the foretell cost is the card's.
//
// Nothing else belongs in the card file either: the face-down exile,
// the owner-only look, the "cast it on a later turn" floor and the
// foretold status on the stack are the keyword's, and the engine
// carries all four.

// Foretell is "Foretell <cost>" — CR 702.143a. `cost` is the printed
// FORETELL COST, the price of casting the card out of exile on a
// later turn: Saw It Coming's "{1}{U}", Cosmic Intervention's
// "{1}{W}{W}".
//
// Give it to the card the way its oracle text reads:
//
//	SpecialActions: []game.SpecialAction{Foretell("{1}{U}")},
//
// A card with foretell and nothing else is otherwise a plain Spec.
func Foretell(cost string) game.SpecialAction {
	return game.SpecialAction{
		Kind:     game.SpecialActionForetell,
		Cost:     game.ForetellExileCost,
		CastCost: cost,
		Label:    "Foretell " + game.ForetellExileCost,
	}
}

// SpellWasForetold reports CR 702.143c — "if this spell was foretold"
// — for the spell currently resolving. Poison the Cup, Haunting
// Voyage and Starnheim Unleashed branch on it.
//
// It is TRUE of a spell cast from a foretold card however that cast
// was paid for, which is why it is not `ctx.PaidAltCost("foretell")`:
// the foretell cost is the usual way to cast a foretold card and not
// the definition of one.
func (c *Context) SpellWasForetold() bool {
	return c.Item != nil && c.Item.Foretold
}

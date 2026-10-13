package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// flash_for_more.go — "You may cast this spell as though it had flash if
// you pay {N} more to cast it" (Asinine Antics, Oakshade Stalker; #3172).
//
// CR 601.3c: an effect that lets a spell be cast as though it had flash
// only if a cost is paid lets its caster BEGIN to cast it at instant
// speed. The engine already has that shape on an alternative cost
// (AlternativeCost.AsThoughFlash, ADR 0129 §5, read by
// CastTimingForOfferOpenLocked in CastSpell, the bot enumerator and the
// view), so the surcharge is one more offer whose price is the printed
// cost plus {N}: claiming it opens the instant-speed window and casting
// for the printed cost does not.
//
// Modelling a surcharge as an alternative cost has one visible edge: the
// offer REPLACES the mana cost, so it cannot be combined with a second
// alternative cost on the same spell (CR 118.9a). No card in the
// catalog that prints this clause has another, and cost reductions
// apply to the offer exactly as they would to the printed cost plus
// {N}, so the price is the same.

// FlashForMore is the offer. `total` is the printed mana cost with the
// surcharge already added ({2}{U}{U} plus {2} is "{4}{U}{U}"), and
// `more` is the surcharge as printed ("{2}") for the label.
//
//	AlternativeCosts: []game.AlternativeCost{FlashForMore("{4}{U}{U}", "{2}")},
func FlashForMore(total, more string) game.AlternativeCost {
	return game.AlternativeCost{
		Key:           "flash-for-more",
		Label:         "Cast with flash — pay " + more + " more (" + total + ")",
		ManaCost:      total,
		AsThoughFlash: true,
	}
}

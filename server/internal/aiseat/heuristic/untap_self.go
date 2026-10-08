package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// untap_self.go — #2500: a self-untap that does not pay for itself.
//
// Basalt Monolith taps for {C}{C}{C} and untaps for {3}. Priced flat
// (ActivateBase, like any activation of the bot's own), "{3}: Untap
// Basalt Monolith" beat passing every time: tap for three, pay three to
// untap, again, with nothing gained. The CR 732 breaker named the loop
// and the runner's #810 hold parked the seat, the same shape #2449 was
// for an Equip {0}.
//
// A row that declares it untaps its own source (`untap_self`) is worth
// taking only when it nets mana: the source's best repeatable mana
// ability, less the row's own cost, is above zero. A Monolith makes 3
// and the untap costs 3 (Basalt) or 4 (Grim), so a round trip makes 0
// or -1, and the bot passes.
//
// This is the general form of #2493's "no gain, no move", for an
// activated row whose net effect is nothing. It does not try to model
// an untap that enables a spell this window: the mana that spell would
// spend is the mana the untap costs, so the source cannot fund it.

// idleSelfUntap reports whether this activation of src is a self-untap
// that nets the bot no mana. Priced idleEquipMove's way, below passing.
func (st *state) idleSelfUntap(src *protocol.CardView, cp activateParams) bool {
	row := rowAt(src, cp.AbilityIndex)
	if row == nil || !row.UntapSelf {
		return false
	}
	return repeatableMana(src)-manaValue(row.ManaCost, 0) <= 0
}

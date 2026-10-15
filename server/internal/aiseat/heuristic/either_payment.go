package heuristic

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"

// either_payment.go — ADR 0187 §7 (#3575): the bot as an offeree of an
// either payment, Worms of the Earth's "any player may sacrifice two
// lands of their choice or have this enchantment deal 5 damage to that
// player. If a player does either, destroy this enchantment."
//
// The offer is one prompt with a move per payable branch (pay_option),
// so the policy scores a branch, not the offer. The order it wants:
//
//  1. Take the damage, when it leaves the seat at or above
//     phyrexianLifeFloor. Life is the cheaper price: a land is a card
//     and a mana a turn, and Worms is the permanent that has stopped
//     every player from replacing one.
//  2. Otherwise sacrifice the least valuable lands, as long as the seat
//     keeps at least eitherLandFloor of them.
//  3. Otherwise decline (0.5, the "no" the caller returns).
//
// The caller has already ruled out paying against the bot's own effect
// and paying for an opponent's permanent that helps the opponent, so
// every branch here is one the bot would like the offer to end on.

// eitherLandFloor is the fewest lands the policy keeps when it
// sacrifices lands to an either payment. Three: below that the seat
// can no longer cast most of what it holds, and the damage was the
// cheaper answer in any case it declined.
const eitherLandFloor = 3

// eitherPaymentValue scores one "pay" move of an either payment. `option`
// is the move's pay_option and `cardIDs` the cards a card branch names.
func (p *Policy) eitherPaymentValue(
	st *state,
	ch *protocol.PendingChoiceView,
	option *int,
	cardIDs []string,
	lookup func(string) *protocol.CardView,
) (float64, string) {
	if option == nil || *option < 0 || *option >= len(ch.PayEither) {
		return 0, "either: no branch named"
	}
	b := ch.PayEither[*option]
	switch b.Action {
	case "damage", "life":
		if st.myLife()-b.Count < phyrexianLifeFloor {
			return phyrexianLifeDeclined, "either: the life would take it below the floor"
		}
		return 0.9, "either: pay with life"
	case "sacrifice":
		// The seat's lands, after the ones this move sacrifices.
		lands := 0
		for i := range st.view.Battlefield.Cards {
			c := &st.view.Battlefield.Cards[i]
			if c.Controller == st.me && isLand(c) {
				lands++
			}
		}
		for _, id := range cardIDs {
			if c := lookup(id); c != nil && isLand(c) {
				lands--
			}
		}
		if lands < eitherLandFloor {
			return 0.25, "either: not enough lands to spare"
		}
		var loss float64
		for _, id := range cardIDs {
			if c := lookup(id); c != nil {
				loss += st.permanentValue(c)
			}
		}
		if loss < 0 {
			loss = 0
		}
		// Above the decline's 0.5, below the damage branch's 0.9: the
		// cheapest lands first, and only when the life is not on offer.
		return 0.85 - 0.3*loss/(1+loss), "either: sacrifice the least valuable lands"
	}
	// A branch the policy has no price for (a discard, an exile, energy)
	// is a payment like any other card or energy cost: take it over the
	// decline, the way a plain payment of that kind is.
	return 0.6, "either: pay"
}

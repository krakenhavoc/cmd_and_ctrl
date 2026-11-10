package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// revealPayment is the one card a reveal / behold cost names (ADR 0100
// amendment 2026-10-07): from the engine's own RevealCostOptionsForEffect,
// so a payment this builds is one validateRevealLocked accepts. The cost
// is the mandatory branch's, or — #4642 — an announced optional behold's
// ("you may behold a Hero"). A beholding seat names a permanent it already
// shows the table when it has one — the options list ends with the
// permanents — and otherwise the first matching card in hand, because
// revealing keeps the card and the policy has nothing to choose between
// two of them. ok is false when the seat has no card to name, which drops
// the branch, or the announced set that includes the optional behold, from
// the move list (it is also unpayable to the view); the unbeholding set is
// walked on its own and stays.
//
// #5406: a behold that chooses its creature type names the type too. ONE
// payment, as for every other cost: the type with the most objects to
// behold (the alphabetically first on a tie), and count() of its objects,
// the permanents first (beholdOrder). A seat with no type it has enough of
// is not offered the set.
func (e *enumerator) revealPayment(castID uuid.UUID, mandatory *game.AdditionalCost, optional []game.AdditionalCost, chosen []int) (ids []uuid.UUID, beholdType string, ok bool) {
	rc := revealCostOf(mandatory, optional, chosen)
	if rc == nil {
		return nil, "", true
	}
	if rc.ChosenType {
		var best *game.BeholdTypeOption
		opts := e.g.BeholdTypeOptionsForEffect(e.seat, castID, rc)
		for i := range opts {
			if best == nil || len(opts[i].Cards) > len(best.Cards) {
				best = &opts[i]
			}
		}
		if best == nil {
			return nil, "", false
		}
		return beholdOrder(e.g, e.seat, best.Cards)[:rc.BeholdCount()], best.Type, true
	}
	ids, ok = e.oneRevealPayment(castID, rc)
	return ids, "", ok
}

// oneRevealPayment is revealPayment's one-card reveal or behold.
func (e *enumerator) oneRevealPayment(castID uuid.UUID, rc *game.RevealCost) (ids []uuid.UUID, ok bool) {
	options := e.g.RevealCostOptionsForEffect(e.seat, castID, rc)
	if len(options) == 0 {
		return nil, false
	}
	if rc.Exile {
		return []uuid.UUID{beholdExilePick(e.g, e.seat, options)}, true
	}
	if rc.Behold {
		return []uuid.UUID{options[len(options)-1]}, true
	}
	return []uuid.UUID{options[0]}, true
}

// beholdExilePick is the one card a seat exiles to a "behold a Kithkin
// and exile it" (#4687), out of the engine's own options. The card is
// spent until the permanent leaves, so the cheapest goes first: a card in
// hand (the one with the lowest mana value), which keeps the seat's board;
// then a token (the one with the lowest power), which is cheap to lose; and
// only then a nontoken permanent, the one with the lowest mana value.
// Ties keep the engine's order.
func beholdExilePick(g *game.Game, seat uuid.UUID, options []uuid.UUID) uuid.UUID {
	p := g.PlayerByIDForEffect(seat)
	rank := func(id uuid.UUID) (tier, n int) {
		c, ok := g.LookupCardForEffect(id)
		switch {
		case !ok:
			return 3, 0
		case p != nil && p.Hand.Contains(id):
			return 0, c.ManaValue()
		case c.IsToken():
			return 1, c.CurrentPower()
		default:
			return 2, c.ManaValue()
		}
	}
	best := options[0]
	bt, bn := rank(best)
	for _, id := range options[1:] {
		if t, n := rank(id); t < bt || (t == bt && n < bn) {
			best, bt, bn = id, t, n
		}
	}
	return best
}

// beholdOrder is an alternative cost's behold candidates (#4689) with the
// permanents first and the hand cards after, each in the engine's order:
// a seat that can behold without revealing its hand does. The options walk
// lists hand cards first, so this only moves the seat's hand to the back.
func beholdOrder(g *game.Game, seat uuid.UUID, pool []uuid.UUID) []uuid.UUID {
	p := g.PlayerByIDForEffect(seat)
	if p == nil {
		return pool
	}
	out := make([]uuid.UUID, 0, len(pool))
	var hand []uuid.UUID
	for _, id := range pool {
		if p.Hand.Contains(id) {
			hand = append(hand, id)
			continue
		}
		out = append(out, id)
	}
	return append(out, hand...)
}

// revealCostOf is the reveal component this announcement pays: the
// mandatory cost's, else the first announced optional cost's. Register
// keeps a card from declaring both (checkOptionalBehold).
func revealCostOf(mandatory *game.AdditionalCost, optional []game.AdditionalCost, chosen []int) *game.RevealCost {
	if mandatory != nil && mandatory.Reveal != nil {
		return mandatory.Reveal
	}
	for _, i := range chosen {
		if i >= 0 && i < len(optional) && optional[i].Reveal != nil {
			return optional[i].Reveal
		}
	}
	return nil
}

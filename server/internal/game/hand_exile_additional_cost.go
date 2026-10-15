package game

import "github.com/google/uuid"

// hand_exile_additional_cost.go — #3175: "As an additional cost to cast
// this spell, you may exile any number of blue cards from your hand.
// This spell costs {2} less to cast for each card exiled this way."
// (March of Swirling Mist and the rest of the Marches.)
//
// An additional cost (CR 601.2b announces it, 601.2f totals the price,
// 601.2h pays it) whose size the caster chooses, from zero up: the
// variable sacrifice's shape (ADR 0100 §3, Dargo's "you may sacrifice any
// number") one zone over, and the exile-a-card cost's move (ExileCost,
// exile_cost.go) on a cast. What it adds is only where the picks ride and
// what reads their number:
//
//   - The caster names the cards on CastSpellParams.ExileFromHandIDs
//     (`exile_from_hand_ids`). None is always a legal payment, so the cost
//     never makes a spell uncastable (CR 118.3: nothing is owed that the
//     caster lacks).
//   - The count is announced before the total is settled, so a modifier
//     reads it as CostQuery.ExilingFromHand — the per-card discount
//     (effects.CostsLessPerExiledFromHand) is part of the price CastSpell
//     charges, the auto-tap preview shows and the enumerator offers, all
//     through the one pricer, as Torgaar's per-sacrifice discount is.
//   - The cards are exiled with the spell already on the stack (CR 601.2a
//     before 601.2h), through the exile-a-card cost's payer: a plain CR
//     406 move, not a discard, so madness never sees it. A countered March
//     does not give them back.

// HandExileCost is "you may exile any number of <kind> cards from your
// hand" as a spell's additional cost. The count is open: zero up to every
// matching card. Nil demands nothing.
type HandExileCost struct {
	// Label is the clause as printed, without the verb — "any number of
	// blue cards from your hand". Shown above the client's picker.
	Label string

	// Match is the clause's predicate, read off the card as it sits in the
	// hand (printed characteristics: a card in a hand has no layers). Nil
	// matches any card. READ-ONLY, under g.mu.
	Match func(Card) bool
}

// matches reports whether `c` could be exiled to this clause. Nil-safe on
// the predicate.
func (h *HandExileCost) matches(c Card) bool {
	return h != nil && (h.Match == nil || h.Match(c))
}

// planHandExile is the hand-exile component the announced plan demands,
// or nil. effects.Register allows it only on the mandatory slot, so there
// is at most one.
func planHandExile(plan []costPayment) *HandExileCost {
	for _, pay := range plan {
		if pay.cost.ExileFromHand != nil {
			return pay.cost.ExileFromHand
		}
	}
	return nil
}

// HandExileOptionsForEffect lists the cards in `playerID`'s hand that may
// be exiled to `cost`, in hand order, without the spell being cast (CR
// 601.2a has moved it to the stack by the time the cost is paid). The
// view ships it and the enumerator pays out of it, so the cards a picker
// offers are the ones the validator accepts (#544).
//
// CALLER MUST ALREADY HOLD g.mu (read or write).
func (g *Game) HandExileOptionsForEffect(playerID, castID uuid.UUID, cost *HandExileCost) []uuid.UUID {
	if cost == nil {
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Hand == nil {
		return nil
	}
	var out []uuid.UUID
	for i := range p.Hand.Cards {
		c := p.Hand.Cards[i]
		if c.InstanceID == castID || !cost.matches(c) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return out
}

// validateCastHandExileLocked checks the cards named to the plan's
// hand-exile cost without moving anything: none when the plan has no such
// component (a list for a spell that charges none is refused rather than
// ignored), else any number of distinct cards in the caster's own hand,
// each matching the clause, none of them the spell, and none also named
// to another component of the same cast (`alsoSpent` — the discards and
// the alternative cost's pitch: one card pays one component, CR 118.3).
//
// Caller must hold g.mu.
func (g *Game) validateCastHandExileLocked(playerID, castID uuid.UUID, plan []costPayment, ids, alsoSpent []uuid.UUID) error {
	cost := planHandExile(plan)
	if cost == nil {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	spent := make(map[uuid.UUID]bool, len(alsoSpent))
	for _, id := range alsoSpent {
		spent[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if id == castID || seen[id] || spent[id] {
			return ErrInvalidParam
		}
		seen[id] = true
		c := findZoneCard(p.Hand, id)
		if c == nil {
			return ErrCardNotFound
		}
		if !cost.matches(*c) {
			return ErrInvalidParam
		}
	}
	return nil
}

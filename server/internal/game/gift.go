package game

import "github.com/google/uuid"

// gift.go — the engine half of gift (CR 702.174, ADR 0089).
//
// Gift is two abilities in one keyword (CR 702.174a). The first is a
// cost: "As an additional cost to cast this spell, you may choose an
// opponent." The second is what the chosen opponent gets. The engine
// owns only the first, because it is the only part that is a CAST
// question — announced at CR 601.2b, validated with the other
// announce-time choices, recorded on the stack item, copied by CR
// 707.10 and carried onto the permanent by CR 400.7d. It rides ADR
// 0073's optional-cost machinery as one more AdditionalCost with
// ChoosesOpponent set, so every one of those steps is the path a
// kicker already takes.
//
// The second ability — the draw, the Food, the tapped Fish — and the
// "if the gift was promised" branches are the card's, grown by
// effects.Gift in the catalog package. Nothing here knows what a Fish
// is.

// GiftPromised reports CR 702.174k for a spell on the stack: its
// controller declared the intention to pay its gift cost, so its gift
// was promised. The one fact every "if the gift was promised" clause
// reads.
func (p PaidCost) GiftPromised() bool { return p.GiftOpponent != uuid.Nil }

// GiftPromised is the same question asked of the permanent a gift
// spell became (CR 400.7d): Scrapshooter's "when this creature
// enters, if the gift was promised" is an intervening if on the
// permanent's own record.
func (p CastProvenance) GiftPromised() bool { return p.GiftOpponent != uuid.Nil }

// GiftPromised on Card for the reason Escaped is there: an
// intervening if (CR 603.4) is handed the permanent and nothing else.
// False for a token, a reanimated card and anything not cast — the
// record is zero for all of them, which is the weaker answer and the
// rule.
func (c Card) GiftPromised() bool { return c.Provenance.GiftPromised() }

// offersGift reports whether any announced optional cost is a gift
// cost — the one that makes CastSpellParams.GiftOpponent mandatory.
func offersGift(optional []AdditionalCost, chosen []int) bool {
	for _, i := range chosen {
		if i >= 0 && i < len(optional) && optional[i].ChoosesOpponent {
			return true
		}
	}
	return false
}

// GiftOpponentsLocked is who `caster` may promise a gift to right now:
// every other seat still in the game, in seat order. CR 702.174a says
// "an opponent", and in a free-for-all Commander game every other
// player is one; a player who has left the game (CR 800.4a) is not a
// player any more and cannot be chosen.
//
// One list for the announce check, the bot enumerator and the view,
// so the client can never offer a recipient the server refuses.
//
// Caller must hold g.mu.
func (g *Game) GiftOpponentsLocked(caster uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, p := range g.Seats {
		if p == nil || p.ID == caster || p.Eliminated {
			continue
		}
		out = append(out, p.ID)
	}
	return out
}

// validateGiftChoiceLocked checks the gift half of the CR 601.2b
// announcement: a gift cost that was announced names an opponent the
// caster may choose, and a gift opponent sent without the cost is
// refused rather than ignored — a client that named a recipient
// believed it was promising a gift, and silently casting the ungifted
// spell is the worst available failure.
//
// Caller must hold g.mu. `chosen` has already passed
// validateOptionalCostChoice.
func (g *Game) validateGiftChoiceLocked(caster uuid.UUID, optional []AdditionalCost, chosen []int, to uuid.UUID) error {
	if !offersGift(optional, chosen) {
		if to != uuid.Nil {
			return ErrInvalidParam
		}
		return nil
	}
	if to == uuid.Nil {
		return ErrInvalidParam
	}
	for _, id := range g.GiftOpponentsLocked(caster) {
		if id == to {
			return nil
		}
	}
	return ErrInvalidParam
}

// optionalCostsDeclareTargets reports whether any of a card's optional
// costs carries a target-clause rewrite — which makes the card one with
// structured targeting even when its printed clause is empty.
func optionalCostsDeclareTargets(optional []AdditionalCost) bool {
	for i := range optional {
		if optional[i].Targets != nil {
			return true
		}
	}
	return false
}

// TargetSpecUnderOptionalCosts applies an optional cost's rewrite of
// the target clause (AdditionalCost.Targets): the first paid optional
// cost that carries one replaces the clause outright, and a cast that
// paid none keeps `base`.
//
// The optional-cost twin of TargetSpecUnderAlternativeCost, read at
// the same three moments — announce (CR 601.2c), resolution (CR
// 608.2b) and the view — so all three judge the spell under the text
// its announcement produced. Applied AFTER the alternative-cost
// rewrite; no printed card offers both a cleave-style rewrite and an
// optional one, and the order is written down so it is a decision
// rather than an accident.
func TargetSpecUnderOptionalCosts(base *TargetSpec, optional []AdditionalCost, paid []int) *TargetSpec {
	for _, i := range paid {
		if i >= 0 && i < len(optional) && optional[i].Targets != nil {
			return optional[i].Targets
		}
	}
	return base
}

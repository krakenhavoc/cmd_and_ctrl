package game

import (
	"github.com/google/uuid"
)

// reveal_cost.go — "reveal a <type> card from your hand" and "behold a
// <type>" as an additional-cost component (CR 701.20, ADR 0100
// amendment 2026-10-07). Both appear only as a branch of an either/or
// cost, paired with mana: Wren's Run Vanquisher's "reveal an Elf card
// from your hand or pay {3}", Silvergill Mentor's "behold a Merfolk or
// pay {2}".
//
// The caster names the ONE card on CastSpellParams.RevealIDs. Paying
// the cost reveals it (RevealForEffect: the whole table sees it, and
// nothing moves — CR 701.20b); a beholding caster may instead name a
// permanent they control, which is chosen and not revealed.

// RevealCost is the reveal / behold component of an additional cost.
type RevealCost struct {
	// Subtype is the creature type the card or permanent must have
	// ("Elf", "Goblin"). A changeling has every creature type in every
	// zone, which Card.HasSubtype already answers.
	Subtype string
	// Behold widens the candidates from "a card in your hand" to "a
	// permanent you control, or a card in your hand" (the keyword
	// action behold, "choose a [quality] you control or reveal a
	// [quality] card from your hand").
	Behold bool
}

// planReveal is the reveal component the announced plan demands, or
// nil. At most one: a cast names one card on RevealIDs.
func planReveal(plan []costPayment) *RevealCost {
	for _, pay := range plan {
		if pay.cost.Reveal != nil {
			return pay.cost.Reveal
		}
	}
	return nil
}

// RevealCostOptionsForEffect lists the cards `playerID` could name to
// pay `rc` when casting `castID`: the matching cards in their hand
// other than the spell itself (CR 601.2a has moved it to the stack by
// the time the cost is paid), then — for behold — the matching
// permanents they control, in battlefield order.
//
// ONE walk, read by the validator, the view and the bot enumerator, so
// the three cannot disagree about who can pay (#544).
//
// Caller must hold g.mu (read or write).
func (g *Game) RevealCostOptionsForEffect(playerID, castID uuid.UUID, rc *RevealCost) []uuid.UUID {
	if rc == nil {
		return nil
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return nil
	}
	var out []uuid.UUID
	for i := range p.Hand.Cards {
		c := &p.Hand.Cards[i]
		if c.InstanceID != castID && c.HasSubtype(rc.Subtype) {
			out = append(out, c.InstanceID)
		}
	}
	if rc.Behold {
		g.RecomputeLayersIfStaleLocked()
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller == playerID && c.HasSubtype(rc.Subtype) {
				out = append(out, c.InstanceID)
			}
		}
	}
	return out
}

// validateRevealLocked checks the card named to pay the plan's reveal
// cost: exactly one when the plan has the component, none otherwise,
// and that one a candidate RevealCostOptionsForEffect lists. Nothing
// is revealed here.
//
// Caller must hold g.mu.
func (g *Game) validateRevealLocked(playerID, castID uuid.UUID, plan []costPayment, ids []uuid.UUID) error {
	rc := planReveal(plan)
	if rc == nil {
		if len(ids) > 0 {
			return ErrInvalidParam
		}
		return nil
	}
	if len(ids) != 1 {
		return ErrInvalidParam
	}
	for _, id := range g.RevealCostOptionsForEffect(playerID, castID, rc) {
		if id == ids[0] {
			return nil
		}
	}
	return ErrInvalidParam
}

// payRevealLocked pays the plan's reveal cost. A card in hand is
// revealed to the table (CR 701.20a: every player sees it, and it
// stays known while it sits there); a permanent chosen to behold is
// not revealed, because choosing it shows nothing the battlefield did
// not. Validated at announce.
//
// Caller must hold g.mu.
func (g *Game) payRevealLocked(playerID, castID uuid.UUID, plan []costPayment, ids []uuid.UUID) {
	rc := planReveal(plan)
	if rc == nil || len(ids) != 1 {
		return
	}
	p := g.playerByIDLocked(playerID)
	if p == nil || !p.Hand.Contains(ids[0]) {
		return
	}
	reason := "reveal a " + rc.Subtype + " card from hand"
	if c, ok := g.LookupCardForEffect(castID); ok {
		reason = c.Name + " — " + reason
	}
	g.RevealForEffect(RevealSpec{Player: playerID, Source: castID, Reason: reason, Cards: ids})
}

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
// An optional "you may behold a <type>" additional cost (#4642, BeholdKey)
// is the same component on an optional entry of the cast's plan: the
// caster who takes it names the card on RevealIDs, and the plan walk below
// finds it wherever it sits.
//
// An alternative cost's "behold N <quality>" (#4689, AlternativeCost.
// Behold: Kindle the Inner Flame's flashback) is the same component with
// a count, named on CastSpellParams.AltCostIDs with the rest of the
// offer's price; validateBeholdPaymentLocked and payBeholdLocked below
// are its validator and payer.
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
	// Color narrows the candidates to one colour ("B", "W", …), read
	// with Card.HasColor — Martyr of Bones' "black cards". Empty means
	// any colour. Combines with Subtype by AND; an additional cost's
	// branch (RevealCardCost) still names a Subtype and no colour.
	// ADR 0020's 2026-10-08 amendment (#2598).
	Color string
	// CardType narrows the candidates to one card type ("creature"), read
	// with Card.HasCardType: a hand card must be a card of that type and a
	// permanent must be one. It is the clause's quality when no creature
	// type is named: Monstrous Emergence's "choose a creature you control
	// or reveal a creature card from your hand" (CR 701.4a, behold's
	// "creature" form). Combines with Subtype and Color by AND. Lower case,
	// as HasCardType takes it.
	CardType string
	// Chosen records the pick on the cast's payment (PaidCost.Chosen), for a
	// spell whose text reads the object it chose or revealed: "damage equal
	// to the power of the creature you chose or the card you revealed". A
	// permanent is recorded as the object it is now, so its power is read as
	// the spell resolves and, once it has left the battlefield, as it last
	// existed there (CR 608.2h); a card from the hand is recorded with the
	// power it was revealed with. Behold only, on the mandatory slot only
	// (ChooseOrRevealCost).
	Chosen bool
	// Exile is "behold a <type> AND EXILE IT" (#4687, ADR 0100's
	// 2026-11-05 amendment): the Lorwyn Eclipsed Champions' mandatory
	// additional cost. The one card named on RevealIDs — a matching
	// permanent the caster controls or a matching card in their hand —
	// is exiled with the spell on the stack (CR 601.2h), linked to the
	// spell object (Card.ExiledWith), and the link moves to the permanent
	// the spell becomes (CR 607.2q, relinkCostExilesLocked), so its
	// "return the exiled card" reads it. Behold only, on the mandatory
	// slot only: effects.Register refuses it anywhere else.
	Exile bool
}

// matches reports whether `c` satisfies the clause's quality: the
// creature type when one is named, the colour when one is named. ONE
// predicate for the cast path's candidate walk and the activation
// path's, so "who can pay" is not written twice (#544).
//
// Both are words of the text of the object whose cost it is (CR 612.2,
// #3984), read through the lens the walk lent `c` (HasSubtypeWord,
// HasColorWord): Martyr of Bones' "Reveal X black cards" under a Sleight
// of Mind is "X red cards". A copy lent nothing reads the printed word.
func (rc *RevealCost) matches(c Card) bool {
	if rc.Subtype != "" && !c.HasSubtypeWord(rc.Subtype) {
		return false
	}
	if rc.Color != "" && !c.HasColorWord(rc.Color) {
		return false
	}
	if rc.CardType != "" && !c.HasCardType(rc.CardType) {
		return false
	}
	return true
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
// The clause's words are read through the text of `castID` while it is
// a permanent (costTextLocked, #3984): an activated ability's reveal
// (the Martyrs) follows a text change of its source. A spell's reveal
// reads its printed words, since its costs are paid before anything can
// change its text.
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
	text := g.costTextLocked(castID)
	var out []uuid.UUID
	for i := range p.Hand.Cards {
		c := &p.Hand.Cards[i]
		if c.InstanceID != castID && rc.matches(lendCostText(text, *c)) {
			out = append(out, c.InstanceID)
		}
	}
	if rc.Behold {
		g.RecomputeLayersIfStaleLocked()
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.Controller == playerID && rc.matches(lendCostText(text, *c)) {
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
// An exiling behold (#4687, RevealCost.Exile) is paid by
// payBeholdExileLocked instead: the card is exiled, not shown.
//
// Caller must hold g.mu.
func (g *Game) payRevealLocked(playerID, castID uuid.UUID, plan []costPayment, ids []uuid.UUID, answers map[uuid.UUID]bool) error {
	rc := planReveal(plan)
	if rc == nil || len(ids) != 1 {
		return nil
	}
	if rc.Exile {
		return g.payBeholdExileLocked(playerID, castID, ids[0], answers)
	}
	if rc.Chosen {
		g.recordChosenLocked(playerID, castID, ids[0])
	}
	p := g.playerByIDLocked(playerID)
	if p == nil || !p.Hand.Contains(ids[0]) {
		return nil
	}
	reason := "reveal a " + rc.Subtype + " card from hand"
	if c, ok := g.LookupCardForEffect(castID); ok {
		reason = c.Name + " — " + reason
	}
	g.RevealForEffect(RevealSpec{Player: playerID, Source: castID, Reason: reason, Cards: ids})
	return nil
}

// validateBeholdPaymentLocked checks the objects named to an alternative
// cost's "behold N <quality>" (#4689, CR 701.4a): exactly n of them, no
// object twice, and each one a candidate RevealCostOptionsForEffect
// lists — a matching card in the caster's hand or a matching permanent
// they control, never the spell being cast. Nothing is revealed here.
//
// Exactly n, like escape's exile: three beholds are three objects, and a
// caster who named the same Elemental three times has beheld one.
//
// Caller must hold g.mu.
func (g *Game) validateBeholdPaymentLocked(playerID, castID uuid.UUID, rc *RevealCost, n int, ids []uuid.UUID) error {
	if len(ids) != n {
		return ErrInvalidParam
	}
	options := make(map[uuid.UUID]bool)
	for _, id := range g.RevealCostOptionsForEffect(playerID, castID, rc) {
		options[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] || !options[id] {
			return ErrInvalidParam
		}
		seen[id] = true
	}
	return nil
}

// payBeholdLocked pays an alternative cost's behold (#4689): the named
// cards in the caster's hand are revealed to the table, as one reveal
// (CR 701.4a, 701.20a), and the chosen permanents are not, because
// choosing a permanent shows nothing the battlefield did not. Nothing
// moves and nothing taps. Validated at announce.
//
// Caller must hold g.mu.
func (g *Game) payBeholdLocked(playerID, castID uuid.UUID, rc *RevealCost, label string, ids []uuid.UUID) {
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return
	}
	var shown []uuid.UUID
	for _, id := range ids {
		if p.Hand.Contains(id) {
			shown = append(shown, id)
		}
	}
	if len(shown) == 0 {
		return
	}
	if label == "" {
		label = rc.Subtype
	}
	reason := "behold " + label
	if c, ok := g.LookupCardForEffect(castID); ok {
		reason = c.Name + " — " + reason
	}
	g.RevealForEffect(RevealSpec{Player: playerID, Source: castID, Reason: reason, Cards: shown})
}

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
	// ChooseCreatureOrWarped is Close Encounter's "As an additional cost to
	// cast this spell, choose a creature you control or a warped creature
	// card you own in exile" (CR 702.185b): the one object named on
	// RevealIDs is a creature the caster controls or a warped creature card
	// they own in exile. Nothing is revealed or moved, so nothing about it
	// is hidden or paid but the choice, and the chosen object rides the
	// spell's payment record (payChosenLocked) for the power it is read for
	// at resolution. Mandatory slot only; Subtype, Color, Behold and Exile
	// are unused with it.
	ChooseCreatureOrWarped bool
	// Count is how many objects one payment beholds: "behold two
	// creatures of that type" is 2 (#5406, Celestial Reunion). Zero is
	// the printed "a" — one. Read through count(). Paid like an
	// alternative cost's behold N (#4689): exactly Count distinct
	// objects on RevealIDs, no object twice.
	Count int `json:",omitempty"`
	// ChosenType is "choose a creature type and behold two creatures of
	// that type" (#5406, CR 601.2b): the clause prints no type, and the
	// caster names one on CastSpellParams.BeholdType as they announce
	// the cost. Subtype is empty on the declared cost; the validator and
	// the payer read the clause bound to the named type (boundTo), and
	// the type lands on the stack item as PaidCost.BeholdType, where the
	// resolution reads it back. On an optional behold only
	// (effects.OptionalBeholdChosenType).
	ChosenType bool `json:",omitempty"`
	// Creature narrows the candidates to creatures: a creature card in
	// hand or a creature you control. Celestial Reunion's "behold two
	// creatures of that type" does not admit a Kindred card or a
	// Kindred permanent of the type (its ruling), so the quality is the
	// card type as well as the creature type.
	Creature bool `json:",omitempty"`
}

// count is how many objects one payment of the clause beholds: Count,
// or one for the printed "a".
func (rc *RevealCost) count() int {
	if rc.Count > 1 {
		return rc.Count
	}
	return 1
}

// Many reports whether a payment names more than one object, or names a
// type as well: the shape validated and paid as a behold N (#5406), not
// as the one-card reveal.
func (rc *RevealCost) Many() bool {
	return rc != nil && (rc.ChosenType || rc.count() > 1)
}

// BeholdCount is how many objects one payment names on reveal_ids.
func (rc *RevealCost) BeholdCount() int {
	if rc == nil {
		return 0
	}
	return rc.count()
}

// boundTo is the clause with the creature type the caster chose at
// announcement written in (#5406). A clause that chooses no type is
// returned as it is.
func (rc *RevealCost) boundTo(subtype string) *RevealCost {
	if rc == nil || !rc.ChosenType {
		return rc
	}
	bound := *rc
	bound.Subtype = subtype
	return &bound
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
	if rc.Creature && !c.IsCreature() {
		return false
	}
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
	if rc.ChooseCreatureOrWarped {
		return g.creatureOrWarpedChoicesLocked(playerID)
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
// #5406: a behold that chooses its creature type (RevealCost.ChosenType)
// also needs `beholdType`, a creature type (CR 205.3m), and then exactly
// count() distinct objects of that type, each a candidate of the clause
// bound to it: validateBeholdPaymentLocked, the alternative cost's behold
// N rule. A type named for a plan that chooses none is refused rather
// than ignored.
//
// Caller must hold g.mu.
func (g *Game) validateRevealLocked(playerID, castID uuid.UUID, plan []costPayment, ids []uuid.UUID, beholdType string) error {
	rc := planReveal(plan)
	if rc == nil {
		if len(ids) > 0 || beholdType != "" {
			return ErrInvalidParam
		}
		return nil
	}
	if rc.ChosenType {
		t, ok := CanonicalCreatureType(beholdType)
		if !ok {
			return ErrInvalidParam
		}
		return g.validateBeholdPaymentLocked(playerID, castID, rc.boundTo(t), rc.count(), ids)
	}
	if beholdType != "" {
		return ErrInvalidParam
	}
	if rc.Many() {
		return g.validateBeholdPaymentLocked(playerID, castID, rc, rc.count(), ids)
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

// planBeholdType is the creature type a validated announcement chose for
// its behold (#5406), in its canonical spelling, or "" when the plan's
// reveal chooses none. The record PaidCost.BeholdType keeps.
func planBeholdType(plan []costPayment, beholdType string) string {
	rc := planReveal(plan)
	if rc == nil || !rc.ChosenType {
		return ""
	}
	t, _ := CanonicalCreatureType(beholdType)
	return t
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
func (g *Game) payRevealLocked(playerID, castID uuid.UUID, plan []costPayment, ids []uuid.UUID, beholdType string, answers map[uuid.UUID]bool) error {
	rc := planReveal(plan)
	if rc == nil || len(ids) == 0 {
		return nil
	}
	// #5406: a behold N, or one of a chosen type — the hand cards are
	// revealed as one reveal, the permanents are chosen.
	if rc.Many() {
		if t := planBeholdType(plan, beholdType); t != "" {
			rc = rc.boundTo(t)
		}
		g.payBeholdLocked(playerID, castID, rc, beholdLabel(rc), ids)
		return nil
	}
	if len(ids) != 1 {
		return nil
	}
	if rc.ChooseCreatureOrWarped {
		g.payChosenLocked(castID, ids[0])
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

// creatureOrWarpedChoicesLocked lists what `playerID` may choose to pay a
// ChooseCreatureOrWarped cost: the creatures they control, in battlefield
// order, then the warped creature cards they own in exile (CR 702.185b), in
// exile order.
//
// Caller must hold g.mu.
func (g *Game) creatureOrWarpedChoicesLocked(playerID uuid.UUID) []uuid.UUID {
	g.RecomputeLayersIfStaleLocked()
	var out []uuid.UUID
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == playerID && c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	for _, c := range g.WarpedCardsOwnedInExileForEffect(playerID) {
		if c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// WarpedCardsOwnedInExileForEffect is the warped cards `owner` owns in exile
// (CR 702.185b): the cards exiled by a warp ability's delayed trigger that are
// still in exile as the same objects. The record is the cast permission warp
// leaves over the card (CastPermission.Warped), which names the object by its
// epoch, so a card that was cast, returned or exiled again is no longer one.
//
// Caller must hold g.mu (read or write).
func (g *Game) WarpedCardsOwnedInExileForEffect(owner uuid.UUID) []Card {
	var out []Card
	for _, c := range g.Exile.Cards {
		if c.Owner != owner {
			continue
		}
		for _, p := range g.Seats {
			if p == nil {
				continue
			}
			warped := false
			for i := range p.CastPermissions {
				perm := &p.CastPermissions[i]
				if perm.Warped && perm.CoversCard(c, ZoneExile) {
					warped = true
					break
				}
			}
			if warped {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

// beholdLabel is the reveal reason's words for a behold N (#5406):
// "two Elf creatures".
func beholdLabel(rc *RevealCost) string {
	n := rc.count()
	words := []string{"", "a", "two", "three", "four", "five"}
	label := rc.Subtype
	if rc.Creature {
		label += " creature"
	}
	if n < len(words) {
		label = words[n] + " " + label
	}
	if n > 1 && rc.Creature {
		label += "s"
	}
	return label
}

// BeholdTypeOption is one creature type a caster could choose for a
// behold that prints none (#5406), with the objects that could pay it.
type BeholdTypeOption struct {
	Type  string
	Cards []uuid.UUID
}

// BeholdTypeOptionsForEffect lists the creature types `playerID` could
// choose to pay `rc` when casting `castID` — a clause with ChosenType —
// each with the objects that could pay it bound to that type, in the
// options walk's order: exactly the types with at least count() of them,
// in CR 205.3m alphabetical order. The types are read off the clause's
// own candidates (a changeling offers every type), and each type's list
// is RevealCostOptionsForEffect of the clause bound to it, so the view,
// the bot and the validator cannot disagree (#544).
//
// Caller must hold g.mu (read or write).
func (g *Game) BeholdTypeOptionsForEffect(playerID, castID uuid.UUID, rc *RevealCost) []BeholdTypeOption {
	if rc == nil || !rc.ChosenType {
		return nil
	}
	types := make(map[string]bool)
	for _, id := range g.RevealCostOptionsForEffect(playerID, castID, rc) {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		for _, t := range CreatureTypesOf(&c) {
			types[t] = true
		}
	}
	var out []BeholdTypeOption
	for _, t := range AllCreatureTypes {
		if !types[t] {
			continue
		}
		if ids := g.RevealCostOptionsForEffect(playerID, castID, rc.boundTo(t)); len(ids) >= rc.count() {
			out = append(out, BeholdTypeOption{Type: t, Cards: ids})
		}
	}
	return out
}

// payChosenLocked records the object a ChooseCreatureOrWarped cost chose on
// the spell's payment record, in the slot a tap-another cost records the
// creature it tapped (PaidCost.TappedOthers: an object, its power when
// chosen, rewritten with its last-known power if it leaves the battlefield,
// CR 608.2h). Close Encounter taps nothing, so the slot is free, and the
// resolution reads it through Context.TappedPower.
//
// Caller must hold g.mu in write mode.
func (g *Game) payChosenLocked(castID, chosen uuid.UUID) {
	item := g.StackMeta[castID]
	if item == nil {
		return
	}
	c, ok := g.LookupCardForEffect(chosen)
	if !ok {
		return
	}
	item.Paid.TappedOthers = paidTapsFrom([]Card{c})
}

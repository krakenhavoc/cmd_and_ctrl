package game

import (
	"sort"

	"github.com/google/uuid"
)

// delve.go — CR 702.66, ADR 0100 §1: "For each generic mana in this
// spell's total cost, you may exile a card from your graveyard rather
// than pay that mana."
//
// Delve is convoke's sibling, one resource over, and it sits next to
// convoke in the pricer for the reason CR 702.66b gives: "The delve
// ability isn't an additional or alternative cost and applies only
// after the total cost of the spell with delve is determined." So it
// is neither an AdditionalCost nor an AlternativeCost. It is a way of
// PAYING the total that costAfterModifiersLocked has already settled:
//
//	printed / alternative cost  → commander tax, optional costs' mana
//	→ cost modifiers (CR 601.2f) → convoke / waterbend taps → delve
//
// It comes after the taps because convoke can pay a coloured symbol
// and delve can pay only a generic one: taking the taps first lets
// convoke's matching put creatures on the coloured symbols, and delve
// then competes only for the generic that is left.
//
// The price is ONE answer (#696). CastPrice.DelveBudget is how many
// cards the announcement may exile, and the validator, the protocol
// view's preview, the bot enumerator and the client's picker all read
// it rather than working a budget out of a cost string.
//
// Delve does not change the spell's mana value (CR 202.3 reads the
// mana cost, and delve is not a cost), and it says nothing about
// CR 107.3b: a free cast of Treasure Cruise has no generic left to
// delve, which the budget reports as zero.

// CatalogDelve is the catalog hook the effects package wires at init
// through CardDef.Delve. Nil, or false, means the card has no delve —
// which is nearly every card.
var CatalogDelve func(key string) bool

// DelveFor reports whether a card being cast has delve, read off the
// catalog entry of the face being cast. A face-down cast (CR 708.4)
// has an empty catalog key and no text, so it never delves.
//
// CR 702.66c makes a second instance of delve on one spell redundant,
// so this is a bool and not a count.
func DelveFor(card Card) bool {
	if CatalogDelve == nil {
		return false
	}
	key := CatalogKey(card)
	if key == "" {
		return false
	}
	return CatalogDelve(key)
}

// CatalogSpellsHaveDelve is the catalog hook for a permanent's "Spells
// you cast have delve" (Teval, Arbiter of Virtue), wired at init
// through CardDef.SpellsYouCastHaveDelve. Nil, or false, means the
// permanent grants nothing.
var CatalogSpellsHaveDelve func(key string) bool

// DelveForLocked is DelveFor for a particular caster. It is the one
// door the cast path, the pricer, the view and the enumerator ask, so
// a GRANTED delve joins here without a second reader (ADR 0100 sub-PR
// 2).
//
// The grant is Teval, Arbiter of Virtue's "Spells you cast have
// delve": a static ability of a permanent on the battlefield (CR
// 113.6, 604.1), read off the caster's permanents at the moment of
// asking and stored nowhere, the way standingCastPermissionsLocked
// derives a permission. It is keyed by CatalogAbilityKey, so a Teval
// that has lost its abilities (CR 613.1f) grants nothing, and a
// permanent that has left grants nothing because it is not there.
// CR 702.66c makes a second source redundant, which is why this stops
// at the first.
//
// A granted delve is a way to PAY, and nothing more. CR 607.2q links
// "exiled with [this object]" only to a delve ability PRINTED on the
// spell, and every card that reads the link prints delve itself, so
// the record the payment leaves (PaidCost.Delved) needs no note of
// where the delve came from.
//
// Caller must hold g.mu.
func (g *Game) DelveForLocked(caster uuid.UUID, card Card) bool {
	if DelveFor(card) {
		return true
	}
	return g.spellsHaveDelveLocked(caster)
}

// spellsHaveDelveLocked reports whether a permanent `caster` controls
// says "Spells you cast have delve".
//
// Caller must hold g.mu.
func (g *Game) spellsHaveDelveLocked(caster uuid.UUID) bool {
	if CatalogSpellsHaveDelve == nil || g.Battlefield == nil || caster == uuid.Nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != caster {
			continue
		}
		if key := CatalogAbilityKey(*c); key != "" && CatalogSpellsHaveDelve(key) {
			return true
		}
	}
	return false
}

// DelvedCardsForEffect resolves CR 607.2q's "cards exiled with [this
// object]" for a delve link (PaidCost.Delved, CastProvenance.Delved,
// PermanentInfo.Delved): the cards that are STILL in exile as the
// object delve put there, in the order the refs name them, as value
// copies.
//
// A ref whose card has left exile is dropped, and so is one whose card
// is in exile again as a NEW object (CR 400.7: it left and came back,
// and nothing links the new object to the spell). That is the whole
// reason the link is a list of ObjectRefs rather than a count, and the
// one place the rule is applied — Murktide Regent's entry counters,
// Soulflayer's keywords and Ethereal Forager's return all read through
// here.
//
// Caller must hold g.mu.
func (g *Game) DelvedCardsForEffect(refs []ObjectRef) []Card {
	if len(refs) == 0 || g.Exile == nil {
		return nil
	}
	out := make([]Card, 0, len(refs))
	for _, ref := range refs {
		for i := range g.Exile.Cards {
			c := g.Exile.Cards[i]
			if c.InstanceID == ref.ID && c.ObjectEpoch == ref.Epoch {
				out = append(out, c)
				break
			}
		}
	}
	return out
}

// delveBudget is how many cards delve may exile against `cost`: the
// generic mana in it, with X folded in at the announced value, and
// WITHOUT the coloured symbols a "spend mana as though it were mana of
// any colour" fold moved into Generic (ParsedCost.FoldedColored). Those
// are still coloured symbols in the rules, and CR 702.66a lets delve
// pay only generic mana — a Breeches-granted Murktide must not delve
// away its {U}{U}. Never negative.
//
// Pure. `cost` is the cost AFTER the modifiers and the taps, because
// the budget is measured against what the cast still owes.
func delveBudget(cost ParsedCost, xValue int) int {
	if xValue < 0 {
		xValue = 0
	}
	n := cost.Generic + cost.XSlots*xValue - cost.FoldedColored
	if n < 0 {
		return 0
	}
	return n
}

// DelveBudgetFor is delveBudget for a reader outside the package that
// already holds the cost after modifiers and taps — the bot
// enumerator, which prices each candidate target set on its own.
func DelveBudgetFor(cost ParsedCost, xValue int) int {
	return delveBudget(cost, xValue)
}

// delveAdjusted subtracts `n` exiled cards from `cost`, capped at the
// budget. X is folded into Generic at the announced value first, the
// way tapPermanentsAdjusted folds it, so the subtraction has one
// concrete number to work against — exactly equivalent for the solver,
// which computes Generic + XSlots*xValue anyway.
//
// n == 0 returns the cost unchanged, {X} slot and all, so a cast that
// delves nothing prices exactly as it did before delve existed.
func delveAdjusted(cost ParsedCost, n, xValue int) ParsedCost {
	if n <= 0 {
		return cost
	}
	if b := delveBudget(cost, xValue); n > b {
		n = b
	}
	out := cost
	out.Required = append([]ColorRequirement(nil), cost.Required...)
	if xValue < 0 {
		xValue = 0
	}
	out.Generic += out.XSlots * xValue
	out.XSlots = 0
	out.Generic -= n
	return out
}

// DelveAdjusted is delveAdjusted for the bot enumerator, which prices
// a candidate payment against a cost it has already put through the
// modifiers.
func DelveAdjusted(cost ParsedCost, n, xValue int) ParsedCost {
	return delveAdjusted(cost, n, xValue)
}

// DelveOptionsForEffect lists the cards `caster` may exile to delve
// the spell `castID`: the cards in the caster's OWN graveyard (CR
// 702.66a says "your graveyard"), except the spell itself — a Hogaak
// cast from the graveyard is on the stack by the time the cost is
// paid (CR 601.2a), but the announcement is validated before it moves
// — and except any card an effect has already paused on its way out
// (#1445: one object pays one cost, CR 118.3).
//
// Returned in the policy-neutral payment order DelvePaymentOrder
// gives, which is the order the view ships the options in, so the
// client's "Choose for me" fills the picker with the cards a bot with
// no fuel policy would have paid (ADR 0100 owner decision 2).
//
// One walk, read by the validator, the view and the enumerator (#544).
//
// Caller must hold g.mu.
func (g *Game) DelveOptionsForEffect(caster, castID uuid.UUID) []uuid.UUID {
	p := g.playerByIDLocked(caster)
	if p == nil || p.Graveyard == nil {
		return nil
	}
	out := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
	for i := range p.Graveyard.Cards {
		c := &p.Graveyard.Cards[i]
		if c.InstanceID == castID {
			continue
		}
		if g.zoneChangePausedLocked(c.InstanceID) {
			continue
		}
		out = append(out, c.InstanceID)
	}
	return g.delvePaymentOrderLocked(p, out)
}

// delvePaymentOrderLocked sorts delve candidates into the order a
// payment takes them from. The order is policy-neutral on purpose, as
// SacrificePaymentOrderForEffect's is: the enumerator re-sorts it by
// the seat's own fuel price when a policy supplies one, and `legal`
// may not import a policy (#687).
//
//  1. lands first — a land does nothing from a graveyard;
//  2. then cards the owner cannot cast from the graveyard before
//     cards they can (flashback, escape, a granted permission), which
//     are the ones worth keeping;
//  3. then the order the graveyard holds them in, oldest first, so the
//     result is stable.
//
// Caller must hold g.mu.
func (g *Game) delvePaymentOrderLocked(p *Player, ids []uuid.UUID) []uuid.UUID {
	type key struct {
		id       uuid.UUID
		land     bool
		castable bool
	}
	keys := make([]key, 0, len(ids))
	for _, id := range ids {
		k := key{id: id}
		for i := range p.Graveyard.Cards {
			c := p.Graveyard.Cards[i]
			if c.InstanceID != id {
				continue
			}
			k.land = c.IsLand()
			k.castable = castableFromOwnGraveyardLocked(g, p, c)
			break
		}
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.land != b.land {
			return a.land
		}
		if a.castable != b.castable {
			return b.castable
		}
		return false
	})
	out := make([]uuid.UUID, len(keys))
	for i, k := range keys {
		out[i] = k.id
	}
	return out
}

// castableFromOwnGraveyardLocked reports whether the graveyard card
// has a cast surface where it sits: its own text opens the graveyard
// (flashback, escape, Gravecrawler), or a permission its owner holds
// does. Used only to ORDER delve candidates; nothing is refused on it.
func castableFromOwnGraveyardLocked(g *Game, p *Player, c Card) bool {
	for _, z := range CastableZonesFor(CatalogKey(c)) {
		if z == ZoneGraveyard {
			return true
		}
	}
	return g.CastPermissionForLocked(p.ID, c, ZoneGraveyard) != nil
}

// validateDelveLocked checks the cards a cast names to delve, without
// exiling anything — the validate-all-then-pay discipline every cost
// component follows, so a rejected cast never leaves a graveyard
// half-emptied.
//
// A card with no delve that arrives WITH delve IDs is a client bug,
// not a no-op, exactly as the additional-cost validator treats a
// stray discard. Naming more cards than the budget is a rejection
// rather than a silent truncation: a card exiled for nothing is a
// card the player lost for nothing. And a card named to delve cannot
// also pay the alternative cost's own graveyard exile (escape), since
// one object pays one cost (CR 118.3).
//
// Caller must hold g.mu.
func (g *Game) validateDelveLocked(caster, castID uuid.UUID, card Card, ids, altCostIDs []uuid.UUID, budget int) error {
	if len(ids) == 0 {
		return nil
	}
	if !g.DelveForLocked(caster, card) {
		return ErrInvalidParam
	}
	if len(ids) > budget {
		return ErrInvalidParam
	}
	allowed := make(map[uuid.UUID]bool)
	for _, id := range g.DelveOptionsForEffect(caster, castID) {
		allowed[id] = true
	}
	alt := make(map[uuid.UUID]bool, len(altCostIDs))
	for _, id := range altCostIDs {
		alt[id] = true
	}
	seen := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		if seen[id] || alt[id] || id == castID {
			return ErrInvalidParam
		}
		seen[id] = true
		if !allowed[id] {
			return ErrCardNotFound
		}
	}
	return nil
}

// payDelveLocked exiles the named cards from their owner's graveyard
// and returns the objects that landed in exile, for PaidCost.Delved.
//
// Called only after validateDelveLocked has passed and after the
// spell is on the stack (CR 601.2a before 601.2h), at the point
// escape's graveyard exile is paid, so anything watching a graveyard
// sees the cards leave while the spell is on the stack — and the
// spell itself, which is not a permanent yet, does not see its own
// delve (Murktide Regent's "whenever an instant or sorcery card leaves
// your graveyard").
//
// Each move is MustSettleNow, because CR 601.2h pays the costs as one
// indivisible step, and carries the CR 903.9 answer its owner gave
// before the cast was paid for (#1397), so a delved commander is
// asked first and never pauses the payment.
//
// The refs name the object IN EXILE — its instance ID and the epoch
// the move gave it (CR 400.7) — because CR 607.2q links a permanent
// only to "cards exiled to pay the cost of the spell that became that
// permanent", and a card that has since left exile is a new object
// that no longer counts.
//
// Caller must hold g.mu in write mode.
func (g *Game) payDelveLocked(ids []uuid.UUID, answers map[uuid.UUID]bool) ([]ObjectRef, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	out := make([]ObjectRef, 0, len(ids))
	for _, id := range ids {
		if _, err := g.routeCardToZoneLocked(zoneRoute{
			CardID:          id,
			Dst:             ZoneExile,
			MustSettleNow:   true,
			commanderAnswer: commanderAnswerFor(answers, id),
		}); err != nil {
			return out, err
		}
		if g.Exile == nil {
			continue
		}
		for i := range g.Exile.Cards {
			if g.Exile.Cards[i].InstanceID == id {
				out = append(out, ObjectRef{ID: id, Epoch: g.Exile.Cards[i].ObjectEpoch})
				break
			}
		}
	}
	return out, nil
}

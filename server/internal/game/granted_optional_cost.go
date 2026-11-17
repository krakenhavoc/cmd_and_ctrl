package game

import "github.com/google/uuid"

// granted_optional_cost.go — #3245: an optional additional cost a
// PERMANENT gives the spells its controller casts. Silverquill, the
// Disputant: "Each instant and sorcery spell you cast has casualty 1."
//
// A card's own optional costs (ADR 0073) are keyed by its oracle ID and
// fixed. A granted one depends on the board at the moment of casting
// (CR 601.2b: the spell has the ability as it is announced), so every
// cast-time reader asks OptionalCostsForCastLocked, which appends the
// costs granted by permanents the caster controls after the card's own.
// The index space the announcement names is therefore "printed, then
// granted", and it is the same list for the offer (legal, protocol) and
// for the validation, pricing and payment in CastSpell.
//
// THE RECORD IS KEYED, NOT INDEXED. A paid granted cost is recorded on
// PaidCost.GrantedOptional by its Key and kept OUT of
// PaidCost.OptionalCosts, whose indices every resolution reader resolves
// against the card's own list. So a kicker reader, a buyback exit or a
// mode count reads exactly what it did before, and a grant whose source
// has left the battlefield since the cast (Silverquill sacrificed to pay
// its own casualty) still reads as paid.
//
// One grant per key: two permanents granting casualty 1 give a spell one
// casualty cost to pay (CR 702.153b's separate instances are not
// modelled; the second copy would need a second sacrifice). A grant
// declared EachSource is the exception (offspring, conspire).
//
// A granted cost that taps permanents is conspire (#4690), whose
// creatures ride the cast's conspire_ids, apart from whatever the spell's
// own cost taps on teamwork_ids (#5688), so it is offered on every spell
// the grant reaches.

// GrantedOptionalCost is one "<spells> you cast have <cost>" grant.
type GrantedOptionalCost struct {
	// Spells says which spells the grant reaches, read against the card
	// being cast as it sits in its zone. Nil is every spell.
	Spells func(g *Game, caster uuid.UUID, spell Card) bool
	// Cost is the optional cost the spell gains. Its Key is the record.
	Cost AdditionalCost
	// Label is the static's own text, for the ability rows.
	Label string
	// EachSource gives the spell one instance of the cost per permanent
	// granting it, instead of one per key (#5428). Offspring is the
	// keyword that needs it: CR 702.175b, "If a spell has multiple
	// instances of offspring, each is paid separately and triggers based
	// on the payments made for it, not any other instances of
	// offspring." Conspire needs it for CR 702.78b, the same sentence
	// (#5688). Each instance is its own entry of the announcement's
	// index space and is recorded under the same Key, so the record
	// counts the payments.
	EachSource bool
}

// PaidLifeForDiscountKey is the key of the optional "pay N life" a
// permanent offers on other spells with a linked discount (#5659, the
// Defiler cycle: "As an additional cost to cast blue permanent spells,
// you may pay 2 life. Those spells cost {U} less to cast if you paid
// life this way."). An on-disk identity (PaidCost.GrantedOptional):
// never renamed, never reused.
const PaidLifeForDiscountKey = "paid_life_for_discount"

// CatalogGrantedOptionalCosts is the catalog hook for a permanent's
// grants, wired in carddef.go from CardDef.GrantedOptionalCosts.
var CatalogGrantedOptionalCosts func(key string) []GrantedOptionalCost

// OptionalCostsForCastLocked is the optional costs `caster` may pay to
// cast `card` now: the card's own (OptionalCostsFor), then one per key
// granted by a permanent `caster` controls, in battlefield order. Each
// granted entry carries Granted, which is how the payment record knows
// to file it by key.
//
// Caller must hold g.mu.
func (g *Game) OptionalCostsForCastLocked(caster uuid.UUID, card Card) []AdditionalCost {
	costs, _ := g.optionalCostsWithGrantorsLocked(caster, card)
	return costs
}

// optionalCostsWithGrantorsLocked is OptionalCostsForCastLocked with,
// beside each entry, the permanent that granted it (uuid.Nil for one of
// the card's own). The second list is nil when nothing was granted.
//
// The grant's Spells filter reads the card through the granting
// permanent's changed words (CR 612.2), so a colour or type word in
// "<kind of> spells you cast" follows a text-changing effect on it.
//
// Caller must hold g.mu.
func (g *Game) optionalCostsWithGrantorsLocked(caster uuid.UUID, card Card) ([]AdditionalCost, []uuid.UUID) {
	own := OptionalCostsFor(CatalogKey(card))
	if CatalogGrantedOptionalCosts == nil || g.Battlefield == nil || caster == uuid.Nil {
		return own, nil
	}
	var out []AdditionalCost
	var grantors []uuid.UUID
	seen := map[string]bool{}
	for _, oc := range own {
		seen[oc.Key] = true
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		if src.Controller != caster {
			continue
		}
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, gr := range CatalogGrantedOptionalCosts(key) {
			if gr.Cost.Key == "" || (seen[gr.Cost.Key] && !gr.EachSource) {
				continue
			}
			if gr.Spells != nil && !gr.Spells(g, caster, ThroughTextOf(src, card)) {
				continue
			}
			seen[gr.Cost.Key] = true
			if out == nil {
				out = append([]AdditionalCost(nil), own...)
				grantors = make([]uuid.UUID, len(own))
			}
			c := gr.Cost
			c.Optional = true
			c.Granted = true
			out = append(out, c)
			grantors = append(grantors, src.InstanceID)
		}
	}
	if out == nil {
		return own, nil
	}
	return out, grantors
}

// GrantedOptionalTimesFromSource is how many times the announcement
// pays the optional cost keyed `key` that q.Source grants the spell
// (#5659): the "if you paid life this way" of Defiler of Dreams' "As an
// additional cost to cast blue permanent spells, you may pay 2 life.
// Those spells cost {U} less to cast if you paid life this way."
//
// The intention to pay is announced at CR 601.2b and the total cost is
// determined at 601.2f, so the reduction reads the announcement
// (CostQuery.OptionalCosts), in the same "printed, then granted" index
// space OptionalCostsForCastLocked gives every other reader. Only the
// entry q.Source itself granted counts: the two sentences are one
// permanent's linked pair, so a second Defiler whose own offer was
// declined reduces nothing. Zero for an activation or special action,
// and for a self modifier (a spell grants itself nothing).
//
// Read-only, under the lock the pricer holds.
func (q CostQuery) GrantedOptionalTimesFromSource(key string) int {
	if q.Game == nil || len(q.OptionalCosts) == 0 || q.Ability != nil || q.SpecialAction != nil || q.Source.InstanceID == uuid.Nil {
		return 0
	}
	costs, grantors := q.Game.optionalCostsWithGrantorsLocked(q.Controller, q.Card)
	n := 0
	for _, i := range q.OptionalCosts {
		if i < 0 || i >= len(grantors) {
			continue
		}
		if costs[i].Granted && costs[i].Key == key && grantors[i] == q.Source.InstanceID {
			n++
		}
	}
	return n
}

// GrantedOptionalTimesPaid counts how many times the granted optional
// cost `key` was paid for the spell whose record is `paid`.
func GrantedOptionalTimesPaid(paid PaidCost, key string) int {
	n := 0
	for _, k := range paid.GrantedOptional {
		if k == key {
			n++
		}
	}
	return n
}

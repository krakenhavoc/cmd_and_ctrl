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
// modelled; the second copy would need a second sacrifice).

// GrantedOptionalCost is one "<spells> you cast have <cost>" grant.
type GrantedOptionalCost struct {
	// Spells says which spells the grant reaches, read against the card
	// being cast as it sits in its zone. Nil is every spell.
	Spells func(g *Game, caster uuid.UUID, spell Card) bool
	// Cost is the optional cost the spell gains. Its Key is the record.
	Cost AdditionalCost
	// Label is the static's own text, for the ability rows.
	Label string
}

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
	own := OptionalCostsFor(CatalogKey(card))
	if CatalogGrantedOptionalCosts == nil || g.Battlefield == nil || caster == uuid.Nil {
		return own
	}
	var out []AdditionalCost
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
			if gr.Cost.Key == "" || seen[gr.Cost.Key] {
				continue
			}
			if gr.Spells != nil && !gr.Spells(g, caster, card) {
				continue
			}
			seen[gr.Cost.Key] = true
			if out == nil {
				out = append([]AdditionalCost(nil), own...)
			}
			c := gr.Cost
			c.Optional = true
			c.Granted = true
			out = append(out, c)
		}
	}
	if out == nil {
		return own
	}
	return out
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

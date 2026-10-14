package game

import "github.com/google/uuid"

// granted_tap_cost.go — a tap-permanents cost (convoke, CR 702.51) a
// PERMANENT gives the spells its controller casts. Hoarding Broodlord:
// "Spells you cast from exile have convoke." Dazzling Theater: "Creature
// spells you cast have convoke." ADR 0073 amendment 2026-10-14.
//
// A card's own convoke is CardDef.TapCost, keyed by its oracle ID. A
// granted one depends on the board and on where the spell is cast from,
// so every cast-time reader asks TapPermanentsCostForCastLocked: the
// card's own cost when it has one, otherwise the first grant from a
// permanent the caster controls whose door (or other designation) gate
// holds and whose Spells filter accepts the card and its zone. Two
// instances of convoke are redundant (CR 702.51b), so a card that prints
// convoke gains nothing more, and a card with a different tap cost of
// its own (waterbend) keeps it.

// GrantedTapCost is one "<spells> you cast have convoke" grant.
type GrantedTapCost struct {
	// Spells says which spells the grant reaches: the card being cast
	// and the zone it is cast from. Nil is every spell.
	Spells func(g *Game, caster uuid.UUID, spell Card, from ZoneKind) bool
	// Cost is the tap cost the spell gains (effects.Convoke()).
	Cost *TapPermanentsCost
	// Label is the static's own text, for the ability rows.
	Label string
	// ActiveWhen gates the grant on a designation of its source (a Room
	// door, CR 709.5). Zero is always on.
	ActiveWhen Designation
}

// CatalogGrantedTapCosts is the catalog hook for a permanent's grants,
// wired in carddef.go from CardDef.GrantedTapCosts.
var CatalogGrantedTapCosts func(key string) []GrantedTapCost

// TapPermanentsCostForCastLocked is the tap cost `caster` may pay to
// cast `card` from `from` now: the card's own, or a granted one.
//
// Caller must hold g.mu.
func (g *Game) TapPermanentsCostForCastLocked(caster uuid.UUID, card Card, from ZoneKind) *TapPermanentsCost {
	if own := TapPermanentsCostFor(CatalogKey(card)); !own.Empty() {
		return own
	}
	if CatalogGrantedTapCosts == nil || g.Battlefield == nil || caster == uuid.Nil {
		return nil
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
		for _, gr := range CatalogGrantedTapCosts(key) {
			if gr.Cost.Empty() {
				continue
			}
			if gr.ActiveWhen.IsGate() && !gr.ActiveWhen.Active(*src) {
				continue
			}
			if gr.Spells != nil && !gr.Spells(g, caster, card, from) {
				continue
			}
			return gr.Cost
		}
	}
	return nil
}

// tapCostForCastParamsLocked is TapPermanentsCostForCastLocked with the
// zone read off a cast's params, for the cast path.
func (g *Game) tapCostForCastParamsLocked(caster uuid.UUID, card Card, params CastSpellParams) *TapPermanentsCost {
	from, _ := castZoneFromWire(params.FromZone)
	return g.TapPermanentsCostForCastLocked(caster, card, from)
}

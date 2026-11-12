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
// card's own cost, a next-spell promise's, and every grant from a
// permanent the caster controls whose door (or other designation) gate
// holds and whose Spells filter accepts the card and its zone. Two
// instances of convoke are redundant (CR 702.51d), so a card that prints
// convoke gains nothing more, and a card with a different tap cost of
// its own (waterbend, improvise) owes both, chained (#5647).

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
// cast `card` from `from` now: the card's own, a promised one and a
// granted one. A spell that has two different ones (its own improvise
// and a granted convoke) owes both, chained (tap_cost_chain.go); a
// second instance of the same keyword adds nothing (CR 702.51d).
//
// Caller must hold g.mu.
func (g *Game) TapPermanentsCostForCastLocked(caster uuid.UUID, card Card, from ZoneKind) *TapPermanentsCost {
	var out *TapPermanentsCost
	if own := TapPermanentsCostFor(CatalogKey(card)); !own.Empty() {
		out = own
	}
	if caster == uuid.Nil {
		return out
	}
	// "The next spell you cast this turn has improvise" (Archway of
	// Innovation) or "has convoke" (Wand of the Worldsoul): a one-use
	// promise on the caster, not a permanent's static.
	out = out.WithAlso(g.promisedTapCostLocked(caster, card))
	if CatalogGrantedTapCosts == nil || g.Battlefield == nil {
		return out
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
			out = out.WithAlso(gr.Cost)
		}
	}
	return out
}

// HasConvoke reports whether the cost, or one chained to it, is convoke
// (CR 702.51). Nil-safe.
func (c *TapPermanentsCost) HasConvoke() bool {
	for _, l := range c.links() {
		if l.Key == "convoke" {
			return true
		}
	}
	return false
}

// SpellHadConvokeForEffect reports whether the spell `spellID` had
// convoke as it became cast (StackItem.HadConvoke): "whenever you cast a
// spell that has convoke" (Joyful Stormsculptor, Kasla, the Broken Halo).
// False for anything that is not a cast spell on the stack.
//
// The *ForEffect surface: caller must hold g.mu.
func (g *Game) SpellHadConvokeForEffect(spellID uuid.UUID) bool {
	item := g.StackMeta[spellID]
	return item != nil && item.HadConvoke
}

// WaterbendTapCost builds a "waterbend {N}" tap cost (effects.Waterbend),
// for a cast permission whose price is a waterbend (#3989). The effects
// package installs it, as it does ImproviseTapCost.
var WaterbendTapCost func(cost string) *TapPermanentsCost

// TapPermanentsCostForGrantLocked is TapPermanentsCostForCastLocked for
// a cast made through `grant`, the permission the cast claims (nil for
// none). A permission priced by waterbending the card's mana value
// (CastPermission.WaterbendManaValue, Hama, the Bloodbender) owes that
// waterbend: its {N} is what CastCostFor charges for the face being
// cast, so the tap cost is marked InCost and the taps pay only that
// {N} (CR 701.67b). The spell's own tap cost, or one the board or a
// promise gives it (convoke, improvise, a waterbend of its own), is
// still owed beside it: the two are chained (tap_cost_chain.go, #4177)
// and the engine shares the one list of taps between them.
// Caller must hold g.mu.
func (g *Game) TapPermanentsCostForGrantLocked(caster uuid.UUID, card Card, from ZoneKind, grant *CastPermission) *TapPermanentsCost {
	own := g.TapPermanentsCostForCastLocked(caster, card, from)
	if grant != nil && grant.WaterbendManaValue && WaterbendTapCost != nil && !card.IsLand() {
		if tc := WaterbendTapCost(WaterbendManaValueCost(card)); !tc.Empty() {
			tc.InCost = true
			return tc.WithAlso(own)
		}
	}
	return own
}

// tapCostForCastParamsLocked is TapPermanentsCostForGrantLocked with the
// zone and the claimed permission read off a cast's params, for the cast
// path: the same CastPermissionForClaimLocked lookup the pricer makes.
//
// #3999: a chosen optional "you may waterbend {N}" cost is the cast's
// tap cost (optional_waterbend.go). effects.Register keeps it off a card
// with a tap cost of its own, so it never displaces one.
func (g *Game) tapCostForCastParamsLocked(caster uuid.UUID, card Card, params CastSpellParams) *TapPermanentsCost {
	from, _ := castZoneFromWire(params.FromZone)
	var grant *CastPermission
	if caster != uuid.Nil && from != ZoneHand && from != ZoneCommand {
		// #4509: judged on the face being cast, as CastSpell judges it,
		// under the permission that opens that face (#4696).
		grant = g.permissionForSpellLocked(g.CastPermissionForFaceLocked(caster, card, from, params.AlternativeCost, params.Face), card)
	}
	own := g.TapPermanentsCostForGrantLocked(caster, card, from, grant)
	if tc := g.optionalWaterbendLocked(caster, card, params.OptionalCosts); tc != nil {
		// A chosen optional waterbend is a cost beside whatever else
		// the cast owes (Hama's waterbend, #4177).
		return tc.WithAlso(own)
	}
	return own
}

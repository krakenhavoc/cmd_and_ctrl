package game

import "github.com/google/uuid"

// ability_zone.go — #660 / ADR 0062 Decision 1: WHERE an activated
// ability functions.
//
// CR 113.6: an ability of a card that is not on the battlefield
// functions only if it says it does. Cycling says so ("Cycling {2}"
// is "{2}, Discard this card: Draw a card", CR 702.29a, and the
// keyword is the saying); Goblin Bombardment does not. Until this
// file the engine had no way to ask: ActivateCatalogAbility looked
// its source up with findBattlefieldCard, so the answer was "the
// battlefield" for every ability on every card.
//
// The dimension is a slice ON THE ABILITY rather than a bool on the
// card, for the reasons ADR 0062 Decision 1 sets out — the next
// cards on the seam row want the graveyard (Reassembling Skeleton,
// Drownyard Temple) and one card wants both (Eternal Dragon's
// plainscycling plus its graveyard return).
//
// The parallel with cast_zones.go is deliberate and worth keeping:
// CastableZones answers "which zone may this card be cast from",
// this answers "which zone does this ability function from", both
// per-declaration, both defaulting to the overwhelmingly common
// answer by leaving the field nil.

// defaultAbilityZones is what an ability that declares nothing gets:
// the battlefield, and nowhere else. That is every ability the
// catalog held before #660.
var defaultAbilityZones = []ZoneKind{ZoneBattlefield}

// AbilityZones is the zones an ability functions from. Never empty.
func AbilityZones(ab ActivatedAbilityShape) []ZoneKind {
	if len(ab.Zones) == 0 {
		return defaultAbilityZones
	}
	return ab.Zones
}

// AbilityFunctionsFromZone reports whether `ab` may be activated
// while its source sits in `zone` (CR 113.6).
//
// Note what it does NOT do: an ability that declares the hand does
// not thereby also work on the battlefield. Cycling is printed on
// lands and on creatures, and a Ketria Triome on the battlefield
// must not offer "Cycling {3}" — there is no card in hand to
// discard, and the paper card cannot be cycled from play either.
// This is the opposite default from CastableZonesFor, which always
// adds the hand, because a cast permission ADDS a path and an
// ability's zone list IS the path.
func AbilityFunctionsFromZone(ab ActivatedAbilityShape, zone ZoneKind) bool {
	for _, z := range AbilityZones(ab) {
		if z == zone {
			return true
		}
	}
	return false
}

// AbilityNeedsPermanentSource reports whether an ability's cost has a
// component that can only be paid by a permanent on the battlefield:
// tapping the source (CR 302.6 summoning sickness rides on it),
// sacrificing the source (CR 701.21a sacrifices a permanent you
// control), crewing (CR 702.122b taps OTHER permanents but reads the
// source as a Vehicle) or a loyalty cost (CR 606.2 puts counters on a
// planeswalker you control).
//
// Used by effects.Register to refuse such a component on an ability
// that declares a non-battlefield zone. The check lives here rather
// than in the catalog because the reason is a rules fact about the
// components, and the components are declared here.
func AbilityNeedsPermanentSource(cost AbilityCost) string {
	switch {
	case cost.Tap:
		return "a tap cost"
	case cost.SacrificeSelf:
		return "a sacrifice-this cost"
	case cost.Crew > 0:
		return "a crew cost"
	case cost.Loyalty != nil:
		return "a loyalty cost"
	}
	return ""
}

// findCardAndZoneLocked finds a card by instance ID and reports which
// zone kind holds it. Mirrors findCardByIDLocked's scan order and
// returns a pointer into the live slice, so the caller may mutate it
// — ActivateCatalogAbility taps its source through this pointer.
//
// The zone comes back as a ZoneKind rather than a *Zone because that
// is what the ability's declaration is written in, and because the
// per-seat zones (hand, library, graveyard, command) are already
// keyed by the owner the caller has.
//
// Caller must hold g.mu.
func (g *Game) findCardAndZoneLocked(cardID uuid.UUID) (*Card, ZoneKind) {
	// #1479: through the card index, not a walk of every zone. The
	// walk this replaced looked at a seat's hand before its library;
	// the index's order differs, which cannot matter while an
	// instance ID is in one zone at a time.
	z, pos := g.locateCardLocked(cardID)
	if z == nil {
		return nil, ""
	}
	return &z.Cards[pos], z.Kind
}

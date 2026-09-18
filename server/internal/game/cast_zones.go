package game

// cast_zones.go — S29: which zone a card may be cast FROM.
//
// Until this file, the answer was a hard-coded whitelist inside
// castSourceZoneLocked: hand, the command zone, and (since S21
// sub-PR 6) exile-with-a-live-impulse-grant. Everything else was
// ErrZoneNotFound, which is why flashback and escape had no path at
// all — the graveyard was not a castable zone for any card, ever.
//
// CR 601.2 casts a spell from wherever the effect that allows it
// says, and the permission is nearly always a property of the CARD:
// "you may cast this card from your graveyard" is printed on
// Gravecrawler, and flashback (CR 702.34b) is the same permission
// with a price attached. So the whitelist becomes a per-card
// declaration — CastableZones — defaulting to hand for the ~99% of
// cards that print nothing.
//
// Two zones stay special and are NOT card properties:
//
//   - The COMMAND zone. CR 903.4 lets a player cast their commander
//     from there regardless of what the card says; the permission
//     belongs to the format, not the card.
//   - EXILE. Impulse exile, airbend, warp and cascade grant
//     permission to one INSTANCE rather than to every copy of the
//     card, so they ride a granted CastPermission (cast_permission.go,
//     ADR 0066). Madness, foretell and suspend, none of them built
//     yet, are the same kind of permission (CR 702.35a, 702.143,
//     702.62a). A card may additionally declare ZoneExile, but that
//     is a CARD-level permission: it opens exile for every copy of
//     the card, at any time, however the copy got there. It fits only
//     a card whose printed text says exactly that, and no catalog
//     card declares it today.
//
//   - THE LIBRARY (S42, CR 401.5). Bolas's Citadel and the Future
//     Sight family open the top card of a library, which is a
//     POSITION rather than a card: no oracle ID can declare it,
//     because the card on top changes every draw. Those are standing
//     CastPermissions derived from the battlefield, and a library is
//     deliberately not declarable in CastableZones.
//
// The PRICE of a non-hand cast is not modelled here. It rides
// AlternativeCost.FromZone, which binds an offer to one zone:
// flashback is an alternative cost claimable only from the
// graveyard, and the two halves compose without either knowing
// about the other.

// CatalogCastableZones is the catalog hook the effects package wires
// at init, mirroring CatalogAlternativeCosts and friends. Nil, or a
// nil return, means the card is castable from hand only — which is
// nearly every card.
var CatalogCastableZones func(oracleID string) []ZoneKind

// defaultCastableZones is what a card that declares nothing gets:
// its owner's hand, and nowhere else (CR 601.1).
var defaultCastableZones = []ZoneKind{ZoneHand}

// CastableZonesFor returns the zones a card declares itself castable
// from. Never empty — a card that declares nothing is castable from
// hand, and a card that declares something does NOT thereby give up
// its hand: flashback adds the graveyard, it does not remove the
// ordinary cast.
func CastableZonesFor(oracleID string) []ZoneKind {
	if CatalogCastableZones == nil || oracleID == "" {
		return defaultCastableZones
	}
	declared := CatalogCastableZones(oracleID)
	if len(declared) == 0 {
		return defaultCastableZones
	}
	out := make([]ZoneKind, 0, len(declared)+1)
	out = append(out, ZoneHand)
	for _, z := range declared {
		if z == ZoneHand {
			continue
		}
		out = append(out, z)
	}
	return out
}

// CardCastableFromZone reports whether the card's own text allows a
// cast out of `zone`. Hand is always true. The command zone and
// exile answer false unless the card says so — their permissions
// come from the format and from per-instance grants respectively,
// and the cast path checks those separately.
func CardCastableFromZone(oracleID string, zone ZoneKind) bool {
	for _, z := range CastableZonesFor(oracleID) {
		if z == zone {
			return true
		}
	}
	return false
}

// castZoneFromWire maps the cast_spell action's `from_zone` string
// onto a ZoneKind. The empty string is hand, which is how every
// pre-S21 client spells "out of my hand" and how the Board's whole
// prompt chain still spells it.
//
// An unrecognised string is a rejection rather than a fall-back to
// hand. The old comment here claimed forward-compatibility as the
// reason for the fall-back, but the fall-back only ever applied to
// values the switch did not list — and casting a card out of the
// wrong zone because the client sent a typo is worse than an error
// toast.
func castZoneFromWire(fromZone string) (ZoneKind, bool) {
	switch fromZone {
	case "", string(ZoneHand):
		return ZoneHand, true
	case string(ZoneCommand):
		return ZoneCommand, true
	case string(ZoneExile):
		return ZoneExile, true
	case string(ZoneGraveyard):
		return ZoneGraveyard, true
	case string(ZoneLibrary):
		// S42, CR 401.5. The wire says "library"; the TOP-CARD half of
		// the rule is enforced by the permission (CastPermission.
		// TopOfLibraryOnly), not by the zone name, because "which
		// pile" and "may you" are separate questions everywhere else
		// in this file too.
		return ZoneLibrary, true
	default:
		return "", false
	}
}

// zoneBoundAlternativeCosts returns the card's alternative costs
// that are claimable only from `zone`. Used by the cast path to
// answer "does casting out of here REQUIRE one of these", and by the
// view layer to decide which offers to show on a card sitting in
// that zone.
func zoneBoundAlternativeCosts(oracleID string, zone ZoneKind) []AlternativeCost {
	var out []AlternativeCost
	for _, ac := range AlternativeCostsFor(oracleID) {
		if ac.FromZone == zone {
			out = append(out, ac)
		}
	}
	return out
}

// AlternativeCostsOfferedFromZone is the view layer's half of the
// same question: everything a caster could claim on a card sitting
// in `zone`.
//
// From hand (and the command zone, which casts a card that is
// otherwise a hand card) that is the unbound offers — overload,
// evoke, cleave. From anywhere else it is exactly the offers bound
// to that zone: a flashback cost is not claimable from hand, and
// overload is not claimable from the graveyard.
func AlternativeCostsOfferedFromZone(oracleID string, zone ZoneKind) []AlternativeCost {
	if zone == ZoneHand || zone == ZoneCommand {
		var out []AlternativeCost
		for _, ac := range AlternativeCostsFor(oracleID) {
			if ac.FromZone == "" || ac.FromZone == ZoneHand {
				out = append(out, ac)
			}
		}
		return out
	}
	return zoneBoundAlternativeCosts(oracleID, zone)
}

// validateCastPathLocked is the S29 gate, widened by ADR 0066: may
// this player cast this card out of this zone, under the alternative
// cost they claimed?
//
// It runs after the alternative cost is resolved (so `alt` is already
// known to be an offer the card makes or the permission synthesises)
// and after the exile permission check. Four rules, in the order they
// fail most often:
//
//  1. A zone-bound offer may only be claimed from its zone. Claiming
//     flashback on a card in hand is a client bug, and charging the
//     flashback cost for a hand cast would be a real rules error in
//     the player's favour.
//  2. A zone neither the card nor a permission opens is not
//     castable — except hand and the command zone, whose permissions
//     come from CR 601.1 and CR 903.4.
//  3. A zone the CARD declares and prices must be paid for. If
//     Faithless Looting offers a graveyard-bound flashback cost, a
//     graveyard cast that claims nothing would otherwise get the
//     printed {R} — strictly better than the card, which is the one
//     direction a sandbox must never err in. Gravecrawler, whose
//     graveyard permission carries no price, declares no bound cost
//     and so is unaffected.
//  4. A PERMISSION that prices its zone must be paid for too, for
//     exactly the same reason: a card Snapcaster gave flashback to
//     is not castable out of the graveyard for its printed cost, and
//     a card on top of the library under Bolas's Citadel is not
//     castable without the life. A permission whose price is a flat
//     Cost (airbend's {2}, cascade's {0}) names no claimable offer
//     and charges itself.
//
// Caller must hold g.mu.
func (g *Game) validateCastPathLocked(card Card, srcKind ZoneKind, alt *AlternativeCost, grant *CastPermission) error {
	key := CatalogKey(card)
	if alt != nil && alt.FromZone != "" && alt.FromZone != srcKind {
		return ErrCastZoneNotAllowed
	}
	switch srcKind {
	case ZoneHand, ZoneCommand:
		return nil
	case ZoneExile:
		// The permission is the usual way in and has already been
		// checked. A card-level ZoneExile declaration is the other;
		// it covers every copy in exile, so it is not the shape for
		// suspend or foretell (see the file header).
		if grant == nil && !CardCastableFromZone(key, ZoneExile) {
			return ErrNoPlayPermission
		}
	default:
		if grant == nil && !CardCastableFromZone(key, srcKind) {
			return ErrCastZoneNotAllowed
		}
	}
	// Rule 4, before rule 3: a permission is the reason this cast is
	// happening at all, so its price is the one that must be paid.
	if grant != nil {
		if offer := grant.AlternativeCostFor(card); offer != nil && alt == nil {
			return ErrCastCostRequired
		}
		return nil
	}
	// Rule 3.
	if bound := zoneBoundAlternativeCosts(key, srcKind); len(bound) > 0 && alt == nil {
		return ErrCastCostRequired
	}
	return nil
}

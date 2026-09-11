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
//   - EXILE. Impulse exile, airbend, and (S29) madness / foretell /
//     suspend all grant permission to one INSTANCE rather than to
//     every copy of the card, so they ride ExilePlayPermission on
//     the card instance instead. A card may additionally declare
//     ZoneExile when the permission really is printed on it.
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

// validateCastPathLocked is the S29 gate: may this player cast this
// card out of this zone, under the alternative cost they claimed?
//
// It runs after validateAlternativeCost (so `alt` is already known
// to be an offer the card makes) and after the ExilePlay check (so
// an exile cast has already proved it holds a grant, or is relying
// on a printed declaration). Three rules, in the order they fail
// most often:
//
//  1. A zone-bound offer may only be claimed from its zone. Claiming
//     flashback on a card in hand is a client bug, and charging the
//     flashback cost for a hand cast would be a real rules error in
//     the player's favour.
//  2. A zone the card does not declare is not castable — except
//     hand, the command zone, and an exile cast riding a live
//     instance grant, all of which are handled by their callers.
//  3. A zone the card declares AND prices must be paid for. If
//     Faithless Looting offers a graveyard-bound flashback cost, a
//     graveyard cast that claims nothing would otherwise get the
//     printed {R} — strictly better than the card, which is the one
//     direction a sandbox must never err in. Gravecrawler, whose
//     graveyard permission carries no price, declares no bound cost
//     and so is unaffected.
//
// Caller must hold g.mu.
func (g *Game) validateCastPathLocked(card Card, srcKind ZoneKind, alt *AlternativeCost, hasExileGrant bool) error {
	key := CatalogKey(card)
	if alt != nil && alt.FromZone != "" && alt.FromZone != srcKind {
		return ErrCastZoneNotAllowed
	}
	switch srcKind {
	case ZoneHand, ZoneCommand:
		return nil
	case ZoneExile:
		// The instance grant is the usual permission and has already
		// been checked. A printed declaration is the other way in.
		if !hasExileGrant && !CardCastableFromZone(key, ZoneExile) {
			return ErrNoPlayPermission
		}
	default:
		if !CardCastableFromZone(key, srcKind) {
			return ErrCastZoneNotAllowed
		}
	}
	// Rule 3. An instance grant carries its own price (airbend's
	// CostOverride) and is not obliged to name a card-level offer.
	if hasExileGrant {
		return nil
	}
	if bound := zoneBoundAlternativeCosts(key, srcKind); len(bound) > 0 && alt == nil {
		return ErrCastCostRequired
	}
	return nil
}

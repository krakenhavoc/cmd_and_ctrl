package game

import "github.com/google/uuid"

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
//     ADR 0066). Foretell (CR 702.143) and suspend (CR 702.62a) are
//     the same kind of permission and shipped as one on #658 / #659;
//     madness (CR 702.35a) will be a third.
//
//     A card may NOT declare ZoneExile. S29 allowed it "for the shape
//     suspend and foretell will use" and #659 retired it, because it
//     is not that shape and could not be: a card-level declaration
//     opens exile for every copy of the card, at any time, however
//     the copy got there, so a Path to Exile'd Rift Bolt would be
//     castable. No catalog card ever declared it, and
//     CastableZonesFor refuses it now so none can.
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

// CastOffersForLocked lists every CR 118.9 cost choice `playerID` may
// announce for `card` out of `zone` right now — the ONE answer to
// "which prices is this cast available at", and the list the bot
// enumerator walks (#673).
//
// A nil entry means the printed mana cost. It is present only when
// the cast path allows a cast that claims nothing, which is what
// keeps a Faithless Looting in the graveyard from being offered at
// the {R} in its corner: a zone the card itself prices must be paid
// for (rule 3 of validateCastPathLocked), and so must a zone a
// permission prices (rule 4).
//
// The rest are offers, in announce precedence: the card's own first,
// then the one a granted permission synthesises, and a granted key
// the card also prints is dropped rather than listed twice — the
// same order resolveAlternativeCostLocked judges a claim in, so a
// Deep Analysis flashed back under Past in Flames is listed at its
// printed price and not at the grant's.
//
// Every entry is filtered through the two gates the announce path
// applies: validateCastPathLocked (is this offer claimable from this
// zone at all) and AlternativeCostPayableLocked (#695 — can its
// condition, CR 119.4's life and CR 601.2b's card component be paid
// right now), which is the same predicate the view's offer stamp
// reads. So a move built on any entry is a move CastSpell accepts,
// which is the enumerator's whole contract — and an EMPTY result is a
// real answer rather than a degenerate one: this card is not castable
// from this zone.
//
// Each entry points at a freshly copied value, so a caller may hold
// one past the call. Caller must hold g.mu.
func (g *Game) CastOffersForLocked(playerID uuid.UUID, card Card, zone ZoneKind, grant *CastPermission) []*AlternativeCost {
	var out []*AlternativeCost
	if g.validateCastPathLocked(card, zone, nil, grant) == nil {
		out = append(out, nil)
	}
	seen := make(map[string]bool, 2)
	add := func(ac *AlternativeCost) {
		if ac == nil || ac.Key == "" || seen[ac.Key] {
			return
		}
		seen[ac.Key] = true
		if g.validateCastPathLocked(card, zone, ac, grant) != nil {
			return
		}
		if !g.AlternativeCostPayableLocked(playerID, card.InstanceID, ac) {
			return
		}
		out = append(out, ac)
	}
	for _, ac := range AlternativeCostsOfferedFromZone(CatalogKey(card), zone) {
		offer := ac
		add(&offer)
	}
	add(grant.AlternativeCostFor(card))
	return out
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
		// A per-instance permission is the ONLY way in, and CastSpell
		// has already checked it. #659 retired the card-level
		// ZoneExile declaration that used to be a second door: it
		// covers every copy in exile, at any time, however the copy
		// got there, which is the wrong shape for suspend, for
		// foretell and for every other exile cast this engine has.
		if grant == nil {
			return ErrNoPlayPermission
		}
	default:
		if grant == nil && !CardCastableFromZone(key, srcKind) {
			return ErrCastZoneNotAllowed
		}
	}
	// Rule 4, and it applies only when the permission is the REASON
	// this cast is legal. A permission must never take away a path the
	// card already prints: Gravecrawler under an Underworld Breach is
	// castable out of the graveyard for free, as it always was, AND
	// for Breach's escape cost, and the caster picks. Demanding the
	// granted price on a card that opens the zone itself would have
	// made a Breach on the table strictly WORSE for its controller,
	// which is the wrong direction twice over.
	if grant != nil && !CardCastableFromZone(key, srcKind) {
		if offer := grant.AlternativeCostFor(card); offer != nil && alt == nil {
			return ErrCastCostRequired
		}
		return nil
	}
	// Rule 3. Reached for a card that opens the zone itself, whether or
	// not a permission also does: the card's own price is still owed.
	if bound := zoneBoundAlternativeCosts(key, srcKind); len(bound) > 0 && alt == nil {
		return ErrCastCostRequired
	}
	return nil
}

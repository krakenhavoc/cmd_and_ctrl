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
// Gravecrawler, and flashback (CR 702.34a) is the same permission
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
// its owner's hand, and nowhere else (CR 601.2).
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

// CardCastableFromAnyFace is CardCastableFromZone asked of EVERY face
// a cast of this card may choose, rather than of the half the pile
// happens to be showing.
//
// One question per face, because an MDFC's halves are SEPARATE
// catalog entries (ADR 0034's "<oracle_id>#N") and only the back may
// print flashback: a card whose back face opens the graveyard has a
// front face whose entry declares nothing, so face 0's answer is "no"
// about a card whose own text says yes.
//
// THE ONE PREDICATE, read by the legal-move enumerator (legal/cast.go)
// and by the view (protocol.stampLegalTargets) rather than each
// spelling the walk out — #1171, where the enumerator asked every face
// and the view asked face 0. The two disagreeing is silent by
// construction: the enumerator offers a bot a cast, the view stamps
// the card with no `castable_here`, no price list and no target
// clause, and the zone browser draws no button behind a cast
// CastSpell would have accepted.
//
// CastableFaces rather than CastableFacesUnder, because both callers
// ask this only when NO permission covers the card. A grant opens its
// zone itself and names its own faces (validateCastPathLocked rule 4,
// faceForCastLocked), so it is a different question with a different
// answer and neither caller reaches here holding one.
func CardCastableFromAnyFace(c Card, zone ZoneKind) bool {
	// The single-faced fast path, which is ~33,000 of the oracle IDs
	// in the game and every card in a typical graveyard: one catalog
	// key, no face walk and no slice.
	if len(c.Faces) < 2 {
		return CardCastableFromZone(CatalogKey(c), zone)
	}
	for _, face := range c.CastableFaces() {
		// Materialised onto a COPY, exactly as CastSpell and the
		// enumerator do it, so CatalogKey reads the chosen half
		// without this function learning how a face becomes a key.
		probe := c
		probe.SetFace(face)
		if CardCastableFromZone(CatalogKey(probe), zone) {
			return true
		}
		// ADR 0103, CR 702.127a: an aftermath half opens its owner's
		// graveyard itself, catalogued or not.
		if opens, _ := aftermathZoneRule(probe, zone); opens {
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
// printed price and not at the grant's. The offers a battlefield static
// applies to every spell the caster casts (ADR 0118 §3: Jodah,
// Omniscience) come last, and one that only repeats a listed price is
// dropped.
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
//
// Deliberately NOT filtered by timing (CR 307.1). This answers "what
// may this cast claim out of this zone", and "is now the right moment"
// is a separate question CastTimingOpenLocked answers on its own —
// castable_here folds both together, and the exile strip's
// informational cast_prices (#1389) reads this list on purpose while a
// sorcery is between windows, so a caster can still see what a card
// will cost later this turn. #1686 needed the picker to know which
// offer is timing-open RIGHT NOW without losing that; see
// AlternativeCostView.TimingClosed and CastSurfaceView.
// PrintedCostTimingClosed, computed alongside this list rather than
// inside it.
func (g *Game) CastOffersForLocked(playerID uuid.UUID, card Card, zone ZoneKind, grant *CastPermission) []*AlternativeCost {
	var out []*AlternativeCost
	if g.validateCastPathLocked(card, zone, nil, grant) == nil {
		out = append(out, nil)
	}
	seen := make(map[string]bool, 2)
	add := func(ac *AlternativeCost, under *CastPermission) {
		if ac == nil || ac.Key == "" || seen[ac.Key] {
			return
		}
		seen[ac.Key] = true
		if g.validateCastPathLocked(card, zone, ac, under) != nil {
			return
		}
		if !g.AlternativeCostPayableLocked(playerID, card.InstanceID, ac) {
			return
		}
		out = append(out, ac)
	}
	for _, ac := range AlternativeCostsOfferedFromZone(CatalogKey(card), zone) {
		offer := ac
		add(&offer, grant)
	}
	// #2167: a permission that opens a zone and charges the PRINTED cost
	// (Muldrotha, an impulse exile) lets the card be cast for any
	// alternative cost it prints instead — Muldrotha's ruling of
	// 2020-11-10, "If it has an alternative cost, you may cast it for that
	// cost instead", which is how a bestow card is cast from the
	// graveyard as an enchantment. The announce path already accepted
	// such a claim (validateCastPathLocked's rule 1 binds only a
	// zone-bound offer); this lists it. A permission with a price of its
	// own ("without paying its mana cost", escape, Bolas's Citadel)
	// lists none: CR 118.9a allows one alternative cost per cast.
	if grant != nil && grant.AltCostKey == "" && grant.Cost == "" && zone != ZoneHand && zone != ZoneCommand {
		for _, ac := range AlternativeCostsOfferedFromZone(CatalogKey(card), ZoneHand) {
			if ac.FromZone != "" {
				continue
			}
			offer := ac
			add(&offer, grant)
		}
	}
	add(grant.AlternativeCostFor(card), grant)
	// #1729: a second stored permission over the same card, priced
	// under an offer of its own (Court of Locthwain's free cast beside
	// its play permission). Judged under that permission, which is the
	// one CastPermissionForClaimLocked hands a cast claiming it.
	for _, other := range g.otherGrantedOffersLocked(playerID, card, zone, grant) {
		add(other.AlternativeCostFor(card), other)
	}
	// ADR 0118 §3, #2163: the offers the caster's permanents grant to
	// every spell they cast (Jodah's {W}{U}{B}{R}{G}, Omniscience's
	// free cast from hand), last, in announce precedence. One that only
	// repeats a price already listed — the printed cost included — is
	// dropped (ADR 0118 call 3). validateCastPathLocked, inside add,
	// keeps each to where the printed mana cost could be paid (CR
	// 118.9a).
	for _, ac := range g.grantedAlternativeCostsLocked(playerID, card, zone) {
		if duplicatesListedPrice(card, out, ac) {
			continue
		}
		add(ac, grant)
	}
	return out
}

// castUsesGrantLocked reports whether a cast of `card` out of `srcKind`
// under `alt` is made BY the permission it was handed, rather than by
// the card's own text — the question a limited permission (CastsLeft,
// #1729) is spent on. It is validateCastPathLocked's "the permission is
// the REASON this cast is legal": exile is opened by nothing else, and
// a graveyard or a library card that opens the zone itself (a printed
// flashback, Gravecrawler, an aftermath half) does not need one.
//
// Caller must hold g.mu.
func (g *Game) castUsesGrantLocked(card Card, srcKind ZoneKind, alt *AlternativeCost) bool {
	switch srcKind {
	case ZoneHand, ZoneCommand:
		return false
	case ZoneExile:
		return true
	}
	aftermathOpens, _ := aftermathZoneRule(card, srcKind)
	return !aftermathOpens && !CardCastableFromZone(castPathKey(card, alt), srcKind)
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
//     come from CR 601.2 and CR 903.4.
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
// An offer a battlefield static GRANTS (ADR 0118 §3, Jodah) is judged
// by one rule of its own instead, CR 118.9a's: it may be claimed only
// where the printed mana cost could be paid. See
// grantedOfferClaimableLocked.
//
// Caller must hold g.mu.
func (g *Game) validateCastPathLocked(card Card, srcKind ZoneKind, alt *AlternativeCost, grant *CastPermission) error {
	if alt != nil && alt.Granted {
		return g.grantedOfferClaimableLocked(card, srcKind, alt, grant)
	}
	// ADR 0107 §4, CR 712.11d: a disturbed card is judged by the zone
	// its FRONT face opens, though the spell is its back face.
	key := castPathKey(card, alt)
	if alt != nil && alt.FromZone != "" && alt.FromZone != srcKind {
		return ErrCastZoneNotAllowed
	}
	// #1665, miracle: an offer that needs a permission to be claimed
	// is claimable only under a live one that names its key. Asked of
	// the grant BEFORE any ForClaim narrowing, because this is the
	// question the grant exists to answer.
	if alt != nil && alt.RequiresGrant && (grant == nil || grant.AltCostKey != alt.Key) {
		return ErrAltCostNotGranted
	}
	// ADR 0099: a discover or cascade grant caps the mana value of the
	// spell it opens, judged against the face the caller materialised —
	// CastSpell and the enumerator both SetFace before asking — so a
	// modal DFC's expensive back face is refused here, in the one gate
	// the view, the bot and the engine share.
	if !spellManaValueWithinCap(card, grant) {
		return ErrSpellManaValueTooHigh
	}
	// ADR 0103, CR 702.127a: an aftermath half casts from a graveyard
	// and from nowhere else — under any permission — and aftermath
	// itself opens the graveyard for that half, and only that half.
	aftermathOpens, err := aftermathZoneRule(card, srcKind)
	if err != nil {
		return err
	}
	opensItself := aftermathOpens || CardCastableFromZone(key, srcKind)
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
		if grant == nil && !opensItself {
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
	if grant != nil && !opensItself {
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

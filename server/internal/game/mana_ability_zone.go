package game

// mana_ability_zone.go — #1228 / CR 113.6, the MANA half of "where
// does this ability function".
//
// ability_zone.go answered the question for CR 602 activated
// abilities in #660 and every consumer of that answer landed with
// #1221. A mana ability is an activated ability (CR 605.1a) but it
// takes the other entry point — ActivateManaAbility, not
// ActivateCatalogAbility — and it carries its cost components on the
// shape rather than in an AbilityCost, so the predicate cannot be
// shared verbatim. What IS shared is every rule underneath it: the
// nil default, the "a declared zone does not also mean the
// battlefield" posture, and the list of components that need a
// permanent to pay them.
//
// Two cards want it, and they are the same card twice:
//
//	Simian Spirit Guide  "Exile this card from your hand: Add {R}."
//	Elvish Spirit Guide  "Exile this card from your hand: Add {G}."
//
// Five consumers read this file, which is one more than the activated
// half has: the activation path (ActivateManaAbility), the legal-move
// enumerator (legal.enumerator.manaMoves), the view's per-seat stamp
// (protocol.stampZoneManaAbilities), the catalog's boot-time refusal
// (effects.Register) — and the AUTO-TAPPER, which has no CR 602
// counterpart at all, because nothing plans a cycling activation on
// the player's behalf and a mana source is exactly the thing the
// planner exists to find.

// defaultManaAbilityZones is what a mana ability that declares
// nothing gets: the battlefield, and nowhere else. That is every mana
// ability in the catalog before #1228, the synthetic basic-land
// ability included.
var defaultManaAbilityZones = []ZoneKind{ZoneBattlefield}

// supportedManaAbilityZones is the set a mana ability may actually
// declare. The HAND alone today, for the reason supportedStaticZones
// is the graveyard alone (#1221): a zone the auto-tapper's gather,
// the enumerator's walk and the view's stamp do not visit would be a
// declaration the engine silently ignores — the card would register,
// look complete on the catalog page, and never produce a mana.
//
// The graveyard, exile and the command zone are each one line here
// plus one pile in gatherManaZoneSources on the day a printed card
// asks. Nothing does today; the whole family is the two Spirit
// Guides.
var supportedManaAbilityZones = []ZoneKind{ZoneHand}

// ManaAbilityZones is the zones a mana ability functions from. Never
// empty.
func ManaAbilityZones(ab ManaAbilityShape) []ZoneKind {
	if len(ab.Zones) == 0 {
		return defaultManaAbilityZones
	}
	return ab.Zones
}

// ManaAbilityFunctionsFromZone reports whether `ab` may be activated
// while its source sits in `zone` (CR 113.6).
//
// Note what it does NOT do, exactly as AbilityFunctionsFromZone does
// not: an ability that declares the hand does not thereby also work
// on the battlefield. A Simian Spirit Guide that somehow reached the
// battlefield is a 2/2 Ape with no abilities, which is what the paper
// card is — the "from your hand" is the ability, not a permission
// added to it.
func ManaAbilityFunctionsFromZone(ab ManaAbilityShape, zone ZoneKind) bool {
	for _, z := range ManaAbilityZones(ab) {
		if z == zone {
			return true
		}
	}
	return false
}

// ManaAbilityDeclaresZone reports whether `ab` names `zone` in its own
// declaration — distinct from ManaAbilityFunctionsFromZone, which
// answers the same question about the DEFAULT too. The boot-time
// checks want this one: "does this ability say it works from a hand"
// is not "does this ability work from a hand by default".
func ManaAbilityDeclaresZone(ab ManaAbilityShape, zone ZoneKind) bool {
	for _, z := range ab.Zones {
		if z == zone {
			return true
		}
	}
	return false
}

// ManaAbilityZoneUnsupported returns why a mana ability may not
// declare `zone`, or "" when it may. The mana sibling of
// StaticZoneUnsupported (#1221), called by effects.Register at boot.
func ManaAbilityZoneUnsupported(zone ZoneKind) string {
	if zone == ZoneBattlefield {
		return ""
	}
	for _, z := range supportedManaAbilityZones {
		if z == zone {
			return ""
		}
	}
	return "only the hand is walked for mana abilities (game.supportedManaAbilityZones); " +
		"a zone nothing walks is a mana ability the engine would silently never offer"
}

// ManaAbilityNeedsPermanentSource reports whether a mana ability's
// cost has a component that only a permanent on the battlefield could
// pay: tapping the source (CR 302.6's sickness rides on it) or
// sacrificing it (CR 701.21a sacrifices a permanent you control).
//
// It is AbilityNeedsPermanentSource with the mana shape's field names
// mapped onto the CR 602 cost's — one rule, two owners, so a
// component that becomes unpayable off the battlefield becomes
// unpayable for both ability kinds at once and neither list can drift
// from the other.
//
// Used by effects.Register to refuse such a component on a mana
// ability that declares a non-battlefield zone, the way it already
// refuses one on an activated ability that does.
func ManaAbilityNeedsPermanentSource(ab ManaAbilityShape) string {
	if why := AbilityNeedsPermanentSource(AbilityCost{
		Tap:           ab.TapCost,
		SacrificeSelf: ab.SacrificeCost,
	}); why != "" {
		return why
	}
	// The counter components, which AbilityCost carries under the
	// same names but which AbilityNeedsPermanentSource does not
	// refuse — because a CR 602 ability may put a counter on a card
	// in EXILE (a suspended card's time counters, CR 702.62a) and
	// this refusal is only ever asked of a non-battlefield zone.
	// Every zone a MANA ability may declare is the hand
	// (supportedManaAbilityZones), and CR 122.6 keeps counters off a
	// card in a hand, so both forms are unpayable here.
	//
	// SacrificeOther and TapOthers are deliberately absent: they name
	// OTHER permanents and a card in hand can name them perfectly
	// well. No printed card does, and the auto-tapper declines both
	// as decisions it may not make, but neither is impossible and
	// this list is for the impossible.
	switch {
	case ab.AddCounter != nil:
		return "a counter-adding cost"
	case ab.RemoveCounters != nil && ab.RemoveCounters.From == nil && !ab.RemoveCounters.Among:
		return "a remove-counters-from-this cost"
	}
	return ""
}

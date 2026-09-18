package game

// landwalk.go — CR 702.14, the evasion ability BlockPairRefusalLocked
// reads in slot 2 after flying (ADR 0045 addendum, Decision 10).
//
// "A creature with landwalk can't be blocked as long as the defending
// player controls at least one land with the specified land type (as in
// "islandwalk"), … without the specified type or supertype (as in
// "nonbasic landwalk") …" (CR 702.14c).
//
// Landwalk is a closed set of canonical keyword tokens, like every
// other keyword the engine enforces: the six below are in
// canonicalKeywords, so the deck importer keeps them off Scryfall's
// keyword array (Cold-Eyed Selkie imports `["Landwalk", "Islandwalk"]`
// as `islandwalk`; the bare "Landwalk" family name is not a token and
// is dropped). The rarer variants CR 702.14a allows — legendary
// landwalk, snow swampwalk, desertwalk, artifact landwalk — join this
// table with the first card that prints one, as the table's
// closedness requires.
//
// Everything is read from EFFECTIVE characteristics, on both sides. A
// granted islandwalk (Lord of Atlantis, layer 6) is visible through
// HasKeyword; a land's types go through IsLand / HasSubtype /
// HasSupertype, which read the post-layer-4 view, so Urborg, Tomb of
// Yawgmoth ("each land is a Swamp in addition to its other land
// types") switches swampwalk on, and a Blood Moon Mountain switches
// mountainwalk on.

import "github.com/google/uuid"

// landwalkSpec is what one landwalk token asks of a land: a land
// subtype it must have, and/or a supertype it must have or (negated)
// must not have.
type landwalkSpec struct {
	subtype   string // "Island"; "" when the token names no land type
	supertype string // "basic"; "" when the token names no supertype
	negated   bool   // the supertype is "without" (nonbasic)
}

// landwalkTokens is every landwalk token the engine enforces, in the
// order the pair check tries them. A slice rather than a map so the
// reported keyword is deterministic when a creature has two.
var landwalkTokens = []struct {
	token string
	spec  landwalkSpec
}{
	{"plainswalk", landwalkSpec{subtype: "Plains"}},
	{"islandwalk", landwalkSpec{subtype: "Island"}},
	{"swampwalk", landwalkSpec{subtype: "Swamp"}},
	{"mountainwalk", landwalkSpec{subtype: "Mountain"}},
	{"forestwalk", landwalkSpec{subtype: "Forest"}},
	{"nonbasic landwalk", landwalkSpec{supertype: "basic", negated: true}},
}

// landwalkRequirement parses a landwalk token into what it asks of a
// land. ok is false for anything that is not an enforced landwalk
// token. The one reader of the table.
func landwalkRequirement(token string) (landwalkSpec, bool) {
	for _, t := range landwalkTokens {
		if t.token == token {
			return t.spec, true
		}
	}
	return landwalkSpec{}, false
}

// matches reports whether `land` is a land with the spec's type and
// supertype, by effective characteristics.
func (s landwalkSpec) matches(land *Card) bool {
	if land == nil || !land.IsLand() {
		return false
	}
	if s.subtype != "" && !land.HasSubtype(s.subtype) {
		return false
	}
	if s.supertype != "" && land.HasSupertype(s.supertype) == s.negated {
		return false
	}
	return true
}

// describe is the land a spec names, with its article, for the
// `illegal_block` sentence: "an Island", "a nonbasic land".
func (s landwalkSpec) describe() string {
	switch {
	case s.subtype != "":
		if s.subtype[0] == 'I' {
			return "an " + s.subtype
		}
		return "a " + s.subtype
	case s.negated:
		return "a non" + s.supertype + " land"
	default:
		return "a " + s.supertype + " land"
	}
}

// landwalkBlockingLandLocked reports the first landwalk ability of
// `attacker` that the defending player's lands switch on, and the land
// that does it. land is nil when no landwalk ability applies — the
// attacker has none, it is not attacking anything (so it has no
// defending player, CR 506.2), or the defending player controls no
// land it names.
//
// Each landwalk ability is checked on its own; two don't cancel
// (CR 702.14d). The defending player is derived from the attack, one
// attacker at a time: the attacked player, a planeswalker's controller,
// or a battle's protector (CR 506.2, 509.1a).
//
// Read-only. Caller holds g.mu (read or write) with fresh layers.
func (g *Game) landwalkBlockingLandLocked(attacker *Card) (string, *Card) {
	if attacker == nil || attacker.AttackingTarget == uuid.Nil {
		return "", nil
	}
	var defender uuid.UUID
	for _, t := range landwalkTokens {
		if !HasKeyword(attacker, t.token) {
			continue
		}
		if defender == uuid.Nil {
			if defender = g.defendingPlayerForAttackLocked(attacker.AttackingTarget); defender == uuid.Nil {
				return "", nil
			}
		}
		for i := range g.Battlefield.Cards {
			land := &g.Battlefield.Cards[i]
			if land.Controller == defender && t.spec.matches(land) {
				return t.token, land
			}
		}
	}
	return "", nil
}

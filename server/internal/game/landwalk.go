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
// is dropped). The rarer variants CR 702.14a allows join this table
// with the first card that prints one, as the table's closedness
// requires: desertwalk and artifact landwalk did, and snow landwalk,
// legendary landwalk and snow <basic land type>walk joined with
// Zombie Musher, Livonya Silone, Rime Dryad and Legions of Lim-Dûl
// (#6143).
//
// One family is not in the table: "snow landwalk of the chosen type"
// (Barbarian Guides) can name any of CR 205.3i's land types, so it is a
// parameterised token, "snow <land type>walk" ("snow gatewalk"), built
// by SnowLandwalkToken and parsed by landwalkRequirement. It is only
// ever granted, never printed, so the importer has nothing to keep and
// it is not canonical; the five basic types are both.
//
// Everything is read from EFFECTIVE characteristics, on both sides. A
// granted islandwalk (Lord of Atlantis, layer 6) is visible through
// HasKeyword; a land's types go through IsLand / HasSubtype /
// HasSupertype, which read the post-layer-4 view, so Urborg, Tomb of
// Yawgmoth ("each land is a Swamp in addition to its other land
// types") switches swampwalk on, and a Blood Moon Mountain switches
// mountainwalk on.

import (
	"strings"

	"github.com/google/uuid"
)

// landwalkSpec is what one landwalk token asks of a land: a land
// subtype it must have, and/or a supertype it must have or (negated)
// must not have.
type landwalkSpec struct {
	subtype   string // "Island"; "" when the token names no land type
	supertype string // "basic"; "" when the token names no supertype
	negated   bool   // the supertype is "without" (nonbasic)
	artifact  bool   // the land must also be an artifact (artifact landwalk)
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
	{"desertwalk", landwalkSpec{subtype: "Desert"}},
	{"nonbasic landwalk", landwalkSpec{supertype: "basic", negated: true}},
	{"artifact landwalk", landwalkSpec{artifact: true}},
	// CR 702.14c: "with the specified type or supertype …, or with both
	// the specified type or supertype and the specified subtype (as in
	// "snow swampwalk")" (#6143).
	{"snow landwalk", landwalkSpec{supertype: "snow"}},
	{"legendary landwalk", landwalkSpec{supertype: "legendary"}},
	{"snow plainswalk", landwalkSpec{subtype: "Plains", supertype: "snow"}},
	{"snow islandwalk", landwalkSpec{subtype: "Island", supertype: "snow"}},
	{"snow swampwalk", landwalkSpec{subtype: "Swamp", supertype: "snow"}},
	{"snow mountainwalk", landwalkSpec{subtype: "Mountain", supertype: "snow"}},
	{"snow forestwalk", landwalkSpec{subtype: "Forest", supertype: "snow"}},
}

// snowLandwalkPrefix and landwalkSuffix frame the parameterised "snow
// <land type>walk" token (Barbarian Guides, #6143).
const (
	snowLandwalkPrefix = "snow "
	landwalkSuffix     = "walk"
)

// SnowLandwalkToken is the keyword token for "snow landwalk of the
// chosen type" (Barbarian Guides): "snow " + the land type in lower
// case + "walk", so a chosen Swamp is the table's own "snow swampwalk"
// and a chosen Gate is "snow gatewalk". landType must be one of CR
// 205.3i's land types; anything else answers "".
func SnowLandwalkToken(landType string) string {
	for _, t := range LandTypes {
		if typeIn(landType, []string{t}) {
			return snowLandwalkPrefix + strings.ToLower(t) + landwalkSuffix
		}
	}
	return ""
}

// LandwalkTokens is every landwalk token the engine reads, the table's
// and the parameterised snow family's for every CR 205.3i land type,
// for an effect that removes "all landwalk abilities" (Hammerheim). A
// fresh slice; the caller may keep it.
func LandwalkTokens() []string {
	out := make([]string, 0, len(landwalkTokens)+len(LandTypes))
	seen := map[string]bool{}
	for _, t := range landwalkTokens {
		out = append(out, t.token)
		seen[t.token] = true
	}
	for _, lt := range LandTypes {
		if tok := SnowLandwalkToken(lt); !seen[tok] {
			out = append(out, tok)
		}
	}
	return out
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
	return parameterisedSnowLandwalk(token)
}

// parameterisedSnowLandwalk parses "snow <land type>walk" for a land
// type outside the table ("snow gatewalk"), the granted-only family
// SnowLandwalkToken builds.
func parameterisedSnowLandwalk(token string) (landwalkSpec, bool) {
	if !strings.HasPrefix(token, snowLandwalkPrefix) || !strings.HasSuffix(token, landwalkSuffix) {
		return landwalkSpec{}, false
	}
	name := strings.TrimSuffix(strings.TrimPrefix(token, snowLandwalkPrefix), landwalkSuffix)
	for _, t := range LandTypes {
		if strings.ToLower(t) == name {
			return landwalkSpec{subtype: t, supertype: "snow"}, true
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
	if s.artifact && !land.IsArtifact() {
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
	case s.artifact:
		return "an artifact land"
	case s.subtype != "" && s.supertype != "":
		return "a " + s.supertype + " " + s.subtype
	case s.subtype != "":
		if strings.ContainsRune("AEIOU", rune(s.subtype[0])) {
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
// or a battle's protector (CR 506.2, 509.1a) — or, once that
// planeswalker or battle has left, the player who was defending it
// (CR 506.4c, 802.2a; #1364), since landwalk is a blocking rule.
//
// Read-only. Caller holds g.mu (read or write) with fresh layers.
func (g *Game) landwalkBlockingLandLocked(attacker *Card) (string, *Card) {
	if attacker == nil || attacker.AttackingTarget == uuid.Nil {
		return "", nil
	}
	var defender uuid.UUID
	noDefender := false
	// check is one landwalk ability against the defending player's
	// lands; it reports the land that switches it on.
	check := func(spec landwalkSpec) *Card {
		if defender == uuid.Nil && !noDefender {
			if defender = g.defendingPlayerForAttackerLocked(attacker); defender == uuid.Nil {
				noDefender = true
			}
		}
		if noDefender {
			return nil
		}
		for i := range g.Battlefield.Cards {
			land := &g.Battlefield.Cards[i]
			if land.Controller == defender && spec.matches(land) {
				return land
			}
		}
		return nil
	}
	for _, t := range landwalkTokens {
		if !HasKeyword(attacker, t.token) || g.landwalkIgnoredLocked(t.token) {
			continue
		}
		if land := check(t.spec); land != nil {
			return t.token, land
		}
		if noDefender {
			return "", nil
		}
	}
	// The parameterised "snow <land type>walk" family (Barbarian
	// Guides), in the attacker's ability order.
	var token string
	var hit *Card
	forEachAbilityToken(attacker, func(a string) bool {
		if !strings.HasPrefix(a, snowLandwalkPrefix) {
			return true
		}
		spec, ok := parameterisedSnowLandwalk(a)
		if !ok || g.landwalkIgnoredLocked(a) {
			return true
		}
		if land := check(spec); land != nil {
			token, hit = a, land
			return false
		}
		return !noDefender
	})
	return token, hit
}

// landwalkIgnoredLocked reports whether a block rule on the table says
// creatures with the landwalk `token` "can be blocked as though they
// didn't have" it (#6195): Quagmire's swampwalk, Staff of the Ages'
// every landwalk ability. Such a creature keeps the keyword; only the
// CR 702.14c evasion is waived, so it is blockable as a creature
// without it would be. Each token is waived on its own name: Quagmire
// waives swampwalk, not snow swampwalk, which is a landwalk ability of
// its own (CR 702.14c, 702.14d).
//
// Read-only. Caller holds g.mu (read or write) with fresh layers.
func (g *Game) landwalkIgnoredLocked(token string) bool {
	ignored := false
	g.forEachBlockRuleLocked(func(r BlockRule, _ *Card) bool {
		if r.IgnoreAllLandwalk {
			ignored = true
			return false
		}
		for _, t := range r.IgnoreLandwalk {
			if t == token {
				ignored = true
				return false
			}
		}
		return true
	})
	return ignored
}

// IsLandwalkToken reports whether token is a landwalk keyword the
// engine enforces ("swampwalk", "nonbasic landwalk"). The one public
// reader of landwalkTokens, for a card that shares or copies "each
// landwalk ability" a creature has (Concerted Effort).
func IsLandwalkToken(token string) bool {
	_, ok := landwalkRequirement(token)
	return ok
}

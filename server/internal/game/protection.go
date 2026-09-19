package game

// protection.go — CR 702.16, the whole of it, and the ONE place in
// the engine that reads a protection quality out of a string.
//
// Protection is the only keyword this engine enforces whose ability
// carries a PARAMETER. Characteristic.Abilities is a []string of bare
// tokens, and ADR 0038 §7 read that as a blocker; it is not, as long
// as the token carries its parameter and exactly one reader parses
// it. That reader is here.
//
//	protection from red
//	protection from artifacts
//	protection from Demons
//	protection from everything
//
// The grammar is CLOSED, in the same sense canonicalKeywords is
// closed: a shape joins it in the change that teaches the engine to
// honour it. A quality this file cannot parse mints NO token, so the
// permanent gets no protection at all and the card keeps its ADR 0037
// unimplemented flag — coverage.go's keywordOnlyLine asks this same
// parser. That errs weaker, which is the only direction this repo
// errs in.
//
// WHAT THE QUALITY IS TESTED AGAINST, and why this file exists at
// all: the SOURCE OBJECT (CR 702.16b-f), never its controller. A red
// player's colourless artifact ability is not red; a white player's
// Lightning Bolt is. Four consumers, one predicate:
//
//	targeting    CanBeTargetedBy      keywords.go   CR 702.16b
//	attachment   attachmentLegalLocked attach.go    CR 702.16c-d
//	damage       protectionPreventsDamageReplacement
//	                                  builtin_replacements.go  CR 702.16e
//	blocking     BlockPairRefusalLocked block_legality.go CR 702.16f
//
// See docs/decisions/0072-protection.md.

import (
	"strings"

	"github.com/google/uuid"
)

// KeywordProtection is the family name. It is NOT a wire token on its
// own: a bare "protection" says nothing about what the permanent is
// protected from, and a badge the rules layer cannot back is the one
// failure mode ADR 0038 exists to prevent. CanonicalKeywords refuses
// it explicitly; the tokens are "protection from <quality>".
const KeywordProtection = "protection"

// protectionPrefix is the token's fixed head, lowercase.
const protectionPrefix = KeywordProtection + " from "

// protectionAndFrom separates the qualities of one printed clause:
// "protection from Demons and from Dragons" is TWO abilities
// (CR 702.16m) and so two tokens.
const protectionAndFrom = " and from "

// ProtectionQualityKind is which characteristic of the source a
// quality is compared against.
type ProtectionQualityKind uint8

const (
	// ProtectionQualityColor compares the source's effective colours.
	// Value is the one-letter wire colour ("W", "U", "B", "R", "G").
	ProtectionQualityColor ProtectionQualityKind = iota
	// ProtectionQualityCardType compares the source's effective card
	// types. Value is the lowercase singular type ("artifact").
	ProtectionQualityCardType
	// ProtectionQualitySubtype compares the source's effective
	// subtypes, and CR 702.73a's AllCreatureTypes flag when the
	// subtype is a creature type — so a changeling really is a Demon
	// to Baneslayer Angel. Value is the canonical singular ("Demon").
	ProtectionQualitySubtype
	// ProtectionQualityEverything is CR 702.16j: every source matches.
	ProtectionQualityEverything
)

// String is the stable wire token for a quality kind, for the
// protocol projection (and through it the client's badge and the
// bot's exchange maths, neither of which may parse a protection
// token itself — ADR 0033 §3 forbids the bot this package at all).
// A token never changes spelling once shipped.
func (k ProtectionQualityKind) String() string {
	switch k {
	case ProtectionQualityColor:
		return "color"
	case ProtectionQualityCardType:
		return "card_type"
	case ProtectionQualitySubtype:
		return "subtype"
	case ProtectionQualityEverything:
		return "everything"
	}
	return ""
}

// ProtectionQuality is one parsed "protection from <quality>".
//
// Printed carries the quality EXACTLY as the token spells it
// ("Demons", "red", "artifacts") because the client's badge tooltip
// shows the quality, and "Protection from Demon" is not what the card
// says. Kind and Value are what the rules read.
type ProtectionQuality struct {
	Kind    ProtectionQualityKind
	Value   string
	Printed string
}

// Token is the canonical wire token this quality was parsed from.
func (q ProtectionQuality) Token() string {
	return protectionPrefix + q.Printed
}

// ParseProtectionQuality parses one whole token, "protection from
// <quality>". Reports false for anything else, including a bare
// "protection", a quality outside the closed grammar, and a clause
// naming two qualities (use ProtectionTokens for printed text).
func ParseProtectionQuality(token string) (ProtectionQuality, bool) {
	t := strings.TrimSpace(token)
	if len(t) <= len(protectionPrefix) || !strings.EqualFold(t[:len(protectionPrefix)], protectionPrefix) {
		return ProtectionQuality{}, false
	}
	return parseQuality(t[len(protectionPrefix):])
}

// parseQuality parses the part after "protection from ". The closed
// grammar, in the order a printed card is most likely to use it.
func parseQuality(raw string) (ProtectionQuality, bool) {
	printed := strings.TrimSpace(raw)
	if printed == "" || strings.Contains(strings.ToLower(printed), protectionAndFrom) {
		return ProtectionQuality{}, false
	}
	lower := strings.ToLower(printed)

	// CR 702.16j. "Protection from everything" is one word and it is
	// checked first because it is not a characteristic lookup at all.
	if lower == "everything" {
		return ProtectionQuality{Kind: ProtectionQualityEverything, Printed: printed}, true
	}
	if col, ok := protectionColors[lower]; ok {
		return ProtectionQuality{Kind: ProtectionQualityColor, Value: col, Printed: printed}, true
	}
	if ct, ok := protectionCardTypes[lower]; ok {
		return ProtectionQuality{Kind: ProtectionQualityCardType, Value: ct, Printed: printed}, true
	}
	// A subtype, and only a CREATURE type: those are the ones printed
	// protection names (Demons, Dragons, Goblins) and the ones
	// CanonicalCreatureType can validate. An artifact or enchantment
	// subtype ("protection from Equipment") joins the grammar with the
	// first card that prints one.
	for _, cand := range singulars(printed) {
		if ct, ok := CanonicalCreatureType(cand); ok {
			return ProtectionQuality{Kind: ProtectionQualitySubtype, Value: ct, Printed: printed}, true
		}
	}
	return ProtectionQuality{}, false
}

// protectionColors maps the printed colour word to the engine's
// one-letter wire colour. The five colours and nothing else:
// "protection from all colors", "from monocolored" and "from
// multicolored" are outside the grammar on purpose (ADR 0072 §10).
var protectionColors = map[string]string{
	"white": "W",
	"blue":  "U",
	"black": "B",
	"red":   "R",
	"green": "G",
}

// protectionCardTypes maps the printed card-type word, singular or
// plural, to the lowercase type HasCardType compares against. CR
// 300.1's list, minus the types no card prints protection from.
var protectionCardTypes = map[string]string{
	"artifact":      "artifact",
	"artifacts":     "artifact",
	"battle":        "battle",
	"battles":       "battle",
	"creature":      "creature",
	"creatures":     "creature",
	"enchantment":   "enchantment",
	"enchantments":  "enchantment",
	"instant":       "instant",
	"instants":      "instant",
	"land":          "land",
	"lands":         "land",
	"planeswalker":  "planeswalker",
	"planeswalkers": "planeswalker",
	"sorcery":       "sorcery",
	"sorceries":     "sorcery",
}

// singulars returns the candidate singular spellings of a printed
// subtype word, most likely first. Protection names its subtypes in
// the plural ("protection from Demons"), and the vocabulary in
// creature_types.go is singular; CanonicalCreatureType validates each
// candidate, so a wrong guess simply fails to match rather than
// inventing a type.
func singulars(word string) []string {
	out := []string{word}
	switch {
	case strings.HasSuffix(word, "ves"):
		// Elves, Dwarves, Wolves.
		out = append(out, word[:len(word)-3]+"f")
	case strings.HasSuffix(word, "s"):
		// Demons, Dragons, Faeries ("Faerie" + "s"), Zombies.
		out = append(out, word[:len(word)-1])
	}
	return out
}

// ProtectionTokens turns ONE printed keyword-ability clause into the
// engine's protection tokens, reporting false when the clause is not
// protection at all or names a quality the grammar cannot parse.
//
// "Protection from Demons and from Dragons" is two abilities
// (CR 702.16m) and so two tokens. A clause whose qualities are not
// ALL parseable yields nothing at all: half a protection is not a
// weaker protection, it is a different card.
//
// This is the parser behind CanonicalKeywords, and through it behind
// the deck importer (deck.printedKeywords) and the ADR 0037 coverage
// signal (coverage.keywordOnlyLine) alike — so what the importer
// stamps and what the badge claims can never drift.
func ProtectionTokens(clause string) ([]string, bool) {
	c := strings.TrimSpace(clause)
	if len(c) <= len(protectionPrefix) || !strings.EqualFold(c[:len(protectionPrefix)], protectionPrefix) {
		return nil, false
	}
	rest := c[len(protectionPrefix):]
	// Split case-insensitively on " and from ", which is how every
	// printed multi-quality protection reads.
	parts := splitAndFrom(rest)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		q, ok := parseQuality(p)
		if !ok {
			return nil, false
		}
		out = append(out, q.Token())
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

// splitAndFrom splits on " and from " case-insensitively, preserving
// each quality's printed spelling.
func splitAndFrom(s string) []string {
	var out []string
	lower := strings.ToLower(s)
	for {
		i := strings.Index(lower, protectionAndFrom)
		if i < 0 {
			out = append(out, s)
			return out
		}
		out = append(out, s[:i])
		s = s[i+len(protectionAndFrom):]
		lower = lower[i+len(protectionAndFrom):]
	}
}

// ProtectionFromColor is the token for "protection from <colour>",
// given the engine's one-letter wire colour ("R"). Empty for anything
// that is not one of the five.
//
// The one place a colour PICK becomes a protection token — the Mother
// of Runes family, whose colour arrives from the choose_color prompt
// (#742) as a wire letter. A card file never spells a token itself,
// for the same reason nothing but this file parses one: the grammar
// has exactly one owner.
func ProtectionFromColor(color string) string {
	for word, letter := range protectionColors {
		if letter == color {
			return protectionPrefix + word
		}
	}
	return ""
}

// ProtectionQualities is THE reader: every protection this card has
// right now, parsed.
//
// Reads the same ability list HasKeyword does — Effective().Abilities
// on the battlefield, so a Sword of Fire and Ice's layer-6 grant
// counts; the card's own printed keywords and then the catalog's
// PrintedKeywords slot off it (CR 113.6: a granted ability does not
// apply outside the battlefield, a printed one is still printed).
//
// Zone-agnostic on purpose. Each consumer applies its own zone rule:
// CanBeTargetedBy gates on the battlefield exactly as ADR 0038 §3
// does for hexproof, and the damage, block and attachment checks are
// battlefield-only by construction.
//
// nil card returns nil.
func ProtectionQualities(c *Card) []ProtectionQuality {
	var out []ProtectionQuality
	forEachAbilityToken(c, func(tok string) bool {
		if q, ok := ParseProtectionQuality(tok); ok {
			out = append(out, q)
		}
		return true
	})
	return out
}

// HasProtection reports whether the card has any protection at all —
// the cheap pre-test every consumer runs before building a source
// snapshot, and the honest answer to the bot's "is this creature
// hard to remove?".
func HasProtection(c *Card) bool {
	found := false
	forEachAbilityToken(c, func(tok string) bool {
		if _, ok := ParseProtectionQuality(tok); ok {
			found = true
			return false
		}
		return true
	})
	return found
}

// Matches reports whether a source with these characteristics has
// the quality (CR 702.16b).
//
// src nil means the source is unknown — a sandbox verb with no source
// card, an effect that names none. Nothing matches an unknown source,
// which lets the damage through and errs weaker; "protection from
// everything" is the one exception, because CR 702.16j names no
// characteristic to look up.
//
// CR 702.16k's player quality ("protection from the chosen player")
// compares Characteristic.Controller rather than a characteristic and
// lands with #929 — which is why the snapshot carries a controller at
// all. See ADR 0072 §7.
func (q ProtectionQuality) Matches(src *Characteristic) bool {
	if q.Kind == ProtectionQualityEverything {
		return true
	}
	if src == nil {
		return false
	}
	switch q.Kind {
	case ProtectionQualityColor:
		return typeListHas(src.Colors, q.Value)
	case ProtectionQualityCardType:
		return typeListHas(src.Types, q.Value)
	case ProtectionQualitySubtype:
		if typeListHas(src.Subtypes, q.Value) {
			return true
		}
		// CR 702.73a via #939's layer-4 flag: a changeling source is
		// every creature type, so it is a Demon to Baneslayer Angel
		// without ~345 entries in Subtypes.
		return src.AllCreatureTypes && IsCreatureType(q.Value)
	}
	return false
}

// ProtectedFrom reports whether `protected` has protection from a
// source with these characteristics — the one predicate all four
// DEBT checks call.
//
// A nil protected card, or one with no protection, returns false.
func ProtectedFrom(protected *Card, src *Characteristic) bool {
	_, ok := MatchedProtection(protected, src)
	return ok
}

// MatchedProtection is ProtectedFrom with the answer: WHICH quality
// refused, in printed order. The player-facing block sentence needs
// it ("Baneslayer Angel has protection from Demons"), and naming the
// quality is the whole reason the token keeps the card's own spelling.
func MatchedProtection(protected *Card, src *Characteristic) (ProtectionQuality, bool) {
	var (
		out   ProtectionQuality
		found bool
	)
	forEachAbilityToken(protected, func(tok string) bool {
		q, ok := ParseProtectionQuality(tok)
		if !ok {
			return true
		}
		if q.Matches(src) {
			out, found = q, true
			return false
		}
		return true
	})
	return out, found
}

// SourceCharacteristics is the source snapshot every consumer that
// holds a live *Card builds: the card's post-layer characteristics
// with its controller materialised, so one value answers both halves
// of CR 702.16 (the quality and, for #929's player quality, who
// controls it).
//
// Returns nil for a nil card, which is how "no source" reaches the
// matcher.
func SourceCharacteristics(c *Card) *Characteristic {
	if c == nil {
		return nil
	}
	ch := c.Effective()
	// Layer 2 only writes Controller on the battlefield; off it, the
	// card's own field is the answer (a spell on the stack is
	// controlled by its caster).
	if ch.Controller == uuid.Nil {
		ch.Controller = c.Controller
	}
	return &ch
}

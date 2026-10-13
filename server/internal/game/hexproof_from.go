package game

// hexproof_from.go — CR 702.11e, "hexproof from [quality]": hexproof
// that only refuses spells and abilities from a source with the
// quality (ADR 0038's amendment of 2026-10-13, #3172).
//
//	hexproof from black
//	hexproof from planeswalkers
//	hexproof from monocolored
//	hexproof from each color
//
// It is the same parameterised-token shape protection is (ADR 0072):
// the token carries its quality, exactly one reader parses it, and the
// quality is tested against the SOURCE of the spell or ability, never
// its caster. The difference from protection is the other half of the
// test. Protection refuses any source with the quality; hexproof from
// refuses only one an OPPONENT controls (CR 702.11b), so the object's
// own controller can still aim a black spell at their own Knight of
// Grace. And it is a targeting restriction only: no damage, enchanting
// or blocking clause (DEBT is protection's, not hexproof's).
//
// The quality grammar REUSES protection's parser for colours, card
// types and creature types, and adds the two colour-count qualities
// only hexproof prints:
//
//	monocolored   the source has exactly one colour
//	each color    the source has at least one colour ("all colors")
//
// Those two stay out of protection's grammar (ADR 0072 §10) because no
// protection ability in the catalog prints them; a quality joins a
// grammar in the change that honours it.
//
// A quality this file cannot parse is no token at all, so the permanent
// has no such ability. That errs weaker, which is the only direction
// this repo errs in; the broad "hexproof" is never substituted (ADR
// 0038 §6 — the 21 cards that print only a narrower hexproof must not
// become fully untargetable).

import (
	"strings"
)

// hexproofFromPrefix is the token's fixed head, lowercase.
const hexproofFromPrefix = KeywordHexproof + " from "

// The colour-count qualities. Lowercase, as the parser compares.
const (
	hexproofQualityMonocolored = "monocolored"
	hexproofQualityEachColor   = "each color"
	hexproofQualityAllColors   = "all colors"
)

// HexproofFromColor is the token for "hexproof from <colour>", given
// the engine's one-letter wire colour ("U"). Empty for anything that
// is not one of the five. The colour-pick counterpart of
// ProtectionFromColor: a card file never spells a token by hand.
func HexproofFromColor(color string) string {
	for word, letter := range protectionColors {
		if letter == color {
			return hexproofFromPrefix + word
		}
	}
	return ""
}

// HexproofFromMonocolored is the token for "hexproof from
// monocolored" (Dragonfire Blade, Sphinx of the Guildpact).
const HexproofFromMonocolored = hexproofFromPrefix + hexproofQualityMonocolored

// HexproofFromEachColor is the token for "hexproof from each color"
// (Breaker of Creation).
const HexproofFromEachColor = hexproofFromPrefix + hexproofQualityEachColor

// HexproofFromPlaneswalkers is the token for "hexproof from
// planeswalkers" (Eradicator Valkyrie).
const HexproofFromPlaneswalkers = hexproofFromPrefix + "planeswalkers"

// IsHexproofFromToken reports whether a token has the shape of a
// "hexproof from <quality>" ability, parseable or not. The cheap
// pre-test every reader runs before parsing.
func IsHexproofFromToken(token string) bool {
	return len(token) > len(hexproofFromPrefix) && strings.EqualFold(token[:len(hexproofFromPrefix)], hexproofFromPrefix)
}

// ParseHexproofFromQuality parses one whole token, "hexproof from
// <quality>". False for anything else, including the bare "hexproof",
// a quality outside the grammar, and the qualities protection has but
// hexproof does not (everything, a player, a mana value, a Ring-bearer
// — none of them is printed as a hexproof variant, so none is
// honoured).
func ParseHexproofFromQuality(token string) (ProtectionQuality, bool) {
	t := strings.TrimSpace(token)
	if !IsHexproofFromToken(t) {
		return ProtectionQuality{}, false
	}
	printed := strings.TrimSpace(t[len(hexproofFromPrefix):])
	switch strings.ToLower(printed) {
	case hexproofQualityMonocolored:
		return ProtectionQuality{Kind: ProtectionQualityMonocolored, Printed: printed}, true
	case hexproofQualityEachColor, hexproofQualityAllColors:
		return ProtectionQuality{Kind: ProtectionQualityAnyColor, Printed: printed}, true
	}
	q, ok := parseQuality(printed)
	if !ok {
		return ProtectionQuality{}, false
	}
	switch q.Kind {
	case ProtectionQualityColor, ProtectionQualityCardType, ProtectionQualitySubtype:
		return q, true
	}
	return ProtectionQuality{}, false
}

// HexproofFromToken is the wire token for a parsed quality.
func HexproofFromToken(q ProtectionQuality) string { return hexproofFromPrefix + q.Printed }

// hexproofFromRefuses reports whether the card has a "hexproof from"
// ability whose quality the spell or ability `src` describes has, and
// the source's controller is an opponent of the card's (CR 702.11b).
// The source snapshot is built at most once, and only for a card that
// has the ability at all.
func hexproofFromRefuses(c *Card, src TargetSource) bool {
	if c == nil || c.Controller == src.Controller {
		return false
	}
	var (
		chars *Characteristic
		read  bool
		out   bool
	)
	forEachAbilityToken(c, func(tok string) bool {
		if !IsHexproofFromToken(tok) {
			return true
		}
		q, ok := ParseHexproofFromQuality(tok)
		if !ok {
			return true
		}
		if !read {
			chars, read = src.Characteristics(), true
		}
		if q.Matches(chars) {
			out = true
			return false
		}
		return true
	})
	return out
}

// HexproofFromTokens returns the card's "hexproof from [quality]"
// tokens the grammar parses, in ability order. For a card that takes
// keywords from another (Soulflayer's "the same is true for hexproof"
// takes the variant too, per its 2020-01-24 ruling).
func HexproofFromTokens(c *Card) []string {
	var out []string
	forEachAbilityToken(c, func(tok string) bool {
		if IsHexproofFromToken(tok) {
			if q, ok := ParseHexproofFromQuality(tok); ok {
				out = append(out, HexproofFromToken(q))
			}
		}
		return true
	})
	return out
}

// HasHexproofFrom reports whether the card has any "hexproof from"
// ability the grammar parses — for readers that ask "is this creature
// hard to remove?".
func HasHexproofFrom(c *Card) bool {
	found := false
	forEachAbilityToken(c, func(tok string) bool {
		if IsHexproofFromToken(tok) {
			if _, ok := ParseHexproofFromQuality(tok); ok {
				found = true
				return false
			}
		}
		return true
	})
	return found
}

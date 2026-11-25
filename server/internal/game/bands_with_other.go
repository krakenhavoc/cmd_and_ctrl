package game

import (
	"strings"

	"github.com/google/uuid"
)

// bands_with_other.go — "bands with other [quality]" (CR 702.22b-c, j-k;
// #6142), the special form of banding (banding.go, ADR 0224).
//
//	702.22b "Bands with other" is a special form of banding. If an effect
//	causes a permanent to lose banding, the permanent loses all "bands
//	with other" abilities as well.
//	702.22c … They may also declare that one or more attacking [quality]
//	creatures with "bands with other [quality]" and any number of other
//	attacking [quality] creatures are all in a band.
//	702.22j … if an attacking creature is being blocked … by both a
//	[quality] creature with "bands with other [quality]" and another
//	[quality] creature, the defending player … chooses how the attacking
//	creature's damage is assigned.
//	702.22k … if a blocking creature is blocking … both a [quality]
//	creature with "bands with other [quality]" and another [quality]
//	creature, the active player … chooses how the blocking creature's
//	damage is assigned.
//
// The ability is an ability token with its quality in it, read off the
// layered ability list like banding, so a printed, granted (Cathedral of
// Serra's "bands with other legendary creatures") or token-carried
// (Wolves of the Hunt's "bands with other creatures named Wolves of the
// Hunt") instance is the same. Two qualities exist on cards:
//
//   - "bands with other legendary creatures": a [quality] creature is a
//     legendary creature (BandsWithOtherLegendary);
//   - "bands with other creatures named <name>": a creature with that
//     name (BandsWithOtherNamed).
//
// A token with any other quality is not read: it bands with nothing and
// hands no damage division over, which errs weaker.
//
// "Loses banding" takes every bands-with-other token with it (CR 702.22b,
// removeKeywordsApply); "loses all 'bands with other' abilities" (Shelkin
// Brownie) is RemoveKeywordPrefixesMod(KeywordBandsWithOtherPrefix).

// KeywordBandsWithOtherPrefix starts every "bands with other" token.
const KeywordBandsWithOtherPrefix = "bands with other "

// BandsWithOtherLegendary is "bands with other legendary creatures"
// (Cathedral of Serra and the Legends lands).
const BandsWithOtherLegendary = KeywordBandsWithOtherPrefix + "legendary creatures"

// bandsWithOtherNamedInfix follows the prefix in the named form.
const bandsWithOtherNamedInfix = "creatures named "

// BandsWithOtherNamed is "bands with other creatures named <name>"
// (Master of the Hunt's Wolves of the Hunt). The name is kept as printed.
func BandsWithOtherNamed(name string) string {
	return KeywordBandsWithOtherPrefix + bandsWithOtherNamedInfix + name
}

// IsBandsWithOther reports whether token is a "bands with other" ability
// the engine reads (one of the two qualities above).
func IsBandsWithOther(token string) bool {
	_, ok := parseBandsWithOther(token)
	return ok
}

// bandQuality is what one "bands with other" token asks of a creature.
type bandQuality struct {
	legendary bool   // "legendary creatures"
	name      string // "creatures named <name>"
}

// parseBandsWithOther reads a token's quality. ok is false for anything
// that is not a "bands with other" token with a quality the engine knows.
func parseBandsWithOther(token string) (bandQuality, bool) {
	if !strings.HasPrefix(token, KeywordBandsWithOtherPrefix) {
		return bandQuality{}, false
	}
	rest := token[len(KeywordBandsWithOtherPrefix):]
	switch {
	case rest == "legendary creatures":
		return bandQuality{legendary: true}, true
	case strings.HasPrefix(rest, bandsWithOtherNamedInfix) && len(rest) > len(bandsWithOtherNamedInfix):
		return bandQuality{name: rest[len(bandsWithOtherNamedInfix):]}, true
	}
	return bandQuality{}, false
}

// admits reports whether c is a [quality] creature, by effective
// characteristics.
func (q bandQuality) admits(c *Card) bool {
	if c == nil || !c.IsCreature() {
		return false
	}
	if q.legendary {
		return c.IsLegendary()
	}
	return PermanentHasName(*c, q.name)
}

// bandsWithOtherQualities is every quality c's own "bands with other"
// tokens name that c itself has: only a [quality] creature WITH "bands
// with other [quality]" anchors a band or a damage hand-off (CR 702.22c,
// j-k).
func bandsWithOtherQualities(c *Card) []bandQuality {
	var out []bandQuality
	forEachAbilityToken(c, func(a string) bool {
		if q, ok := parseBandsWithOther(a); ok && q.admits(c) {
			out = append(out, q)
		}
		return true
	})
	return out
}

// bandsWithOtherBandLocked reports whether the creatures `ids` may form
// one band by CR 702.22c's second sentence: some member is a [quality]
// creature with "bands with other [quality]" and every member is a
// [quality] creature. Layers must be fresh.
//
// Caller must hold g.mu.
func (g *Game) bandsWithOtherBandLocked(ids []uuid.UUID) bool {
	cards := make([]*Card, 0, len(ids))
	for _, id := range ids {
		if c := findBattlefieldCard(g, id); c != nil {
			cards = append(cards, c)
		}
	}
	for _, anchor := range cards {
		for _, q := range bandsWithOtherQualities(anchor) {
			all := true
			for _, c := range cards {
				if !q.admits(c) {
					all = false
					break
				}
			}
			if all {
				return true
			}
		}
	}
	return false
}

// bandsWithOtherPairLocked reports whether `ids` hold both a [quality]
// creature with "bands with other [quality]" and ANOTHER [quality]
// creature (CR 702.22j-k), read as the damage is assigned.
//
// Caller must hold g.mu.
func (g *Game) bandsWithOtherPairLocked(ids []uuid.UUID) bool {
	for _, id := range ids {
		anchor := findBattlefieldCard(g, id)
		for _, q := range bandsWithOtherQualities(anchor) {
			for _, other := range ids {
				if other != id && q.admits(findBattlefieldCard(g, other)) {
					return true
				}
			}
		}
	}
	return false
}

// hasAnyBandsWithOther reports whether c has a "bands with other" token
// the engine reads.
func hasAnyBandsWithOther(c *Card) bool {
	found := false
	forEachAbilityToken(c, func(a string) bool {
		if IsBandsWithOther(a) {
			found = true
			return false
		}
		return true
	})
	return found
}

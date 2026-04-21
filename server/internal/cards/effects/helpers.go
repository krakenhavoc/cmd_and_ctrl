package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// IsBasicLand reports whether a card's type line contains the
// "basic land" supertype (case-insensitive substring). Used by
// tutor / fetch primitives that need to match Forest / Island /
// Mountain / Plains / Swamp / Wastes without enumerating each
// subtype. Path to Exile, Cultivate, and Solemn Simulacrum share
// this predicate; keep it here rather than duplicating the
// lowercase loop per-card.
func IsBasicLand(c game.Card) bool {
	return containsFoldASCII(c.TypeLine, "basic land")
}

// containsFoldASCII is a zero-alloc case-insensitive substring
// check for ASCII-only needles. The type lines we feed this come
// straight from Scryfall and are plain ASCII; if future text goes
// Unicode the caller should switch to strings.EqualFold + slicing.
func containsFoldASCII(haystack, needle string) bool {
	if len(haystack) < len(needle) {
		return false
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			h := haystack[i+j]
			if h >= 'A' && h <= 'Z' {
				h += 'a' - 'A'
			}
			if h != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

package game

import "strings"

// lose_keyword_families.go — two scoped records for "permanents your
// opponents control lose hexproof, indestructible, protection, shroud, and
// ward until end of turn" (Shay Cormac, #5948). Hexproof, indestructible and
// shroud are exact tokens (ModRemoveKeywords); what was left is the two
// kinds of keyword that are not one token:
//
//   - protection is a family of tokens, "protection from <quality>" (CR
//     702.16), and "hexproof from <quality>" (CR 702.11e) is another, so a
//     permanent that loses protection loses every one of them:
//     ModRemoveKeywordPrefixes drops each ability token that starts with one
//     of the prefixes, in layer 6 like any other removal (CR 613.1f).
//   - ward (CR 702.21) is a triggered ability, never an ability token, so
//     losing it is the ward trigger not happening: ModSuppressWard is read
//     by WardSuppressedForEffect, the gate every ward in the catalog is built
//     on, for the permanents the record names.
//
// Both are additive record kinds. Mod kinds are on-disk identities: never
// renamed, never reused.

const (
	// ModRemoveKeywordPrefixes removes every ability token that starts with
	// one of Keywords (compared without case): "protection from " and
	// "hexproof from ". Layer 6, a CR 613.1f removal.
	ModRemoveKeywordPrefixes ModKind = "removeKeywordPrefixes"
	// ModSuppressWard is "<affected> lose ward": the ward abilities of the
	// affected permanents don't trigger. Reads nothing; not a layer
	// operation (WardSuppressedForEffect reads it).
	ModSuppressWard ModKind = "suppressWard"
)

// RemoveKeywordPrefixesMod is "loses every <prefix…> ability" as a data
// record. Reads Keywords, which must not be empty.
func RemoveKeywordPrefixesMod(prefixes ...string) Mod {
	return Mod{Kind: ModRemoveKeywordPrefixes, Keywords: copyStrings(prefixes)}
}

// SuppressWardMod is "loses ward" as a data record.
func SuppressWardMod() Mod { return Mod{Kind: ModSuppressWard} }

// removeKeywordPrefixesApply is the layer-6 Apply of ModRemoveKeywordPrefixes.
func removeKeywordPrefixesApply(prefixes []string) func(*Characteristic, *Card, *Game, *Card) {
	return func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
		kept := make([]string, 0, len(ch.Abilities))
		for _, a := range ch.Abilities {
			drop := false
			for _, p := range prefixes {
				if len(a) >= len(p) && strings.EqualFold(a[:len(p)], p) {
					drop = true
					break
				}
			}
			if !drop {
				kept = append(kept, a)
			}
		}
		ch.Abilities = kept
	}
}

// loseKeywordFamiliesModProblem is the registration check for the two kinds:
// "" when the mod is sound.
func loseKeywordFamiliesModProblem(m Mod) string {
	switch m.Kind {
	case ModRemoveKeywordPrefixes:
		if len(m.Keywords) == 0 {
			return "a removeKeywordPrefixes mod names no prefix"
		}
	case ModSuppressWard:
		if len(m.Keywords) != 0 {
			return "a suppressWard mod reads no keywords"
		}
	}
	return ""
}

// scopedWardSuppressedLocked reports whether a live record says the ward
// abilities of `warded` don't trigger. Like every record that changes
// abilities, its set is the permanents it found as it resolved (CR 611.2c).
// Caller must hold g.mu. Reads only.
func (g *Game) scopedWardSuppressedLocked(warded *Card) bool {
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		if !scopedEffectHasMod(e, ModSuppressWard) {
			continue
		}
		applies := affectedPredicate(e.Affected)
		if e.Scope != ScopeNone {
			applies = scopePredicate(e.Scope, e.Controller)
		}
		if applies(warded, g, nil) {
			return true
		}
	}
	return false
}

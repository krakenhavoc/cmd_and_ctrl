package game

import (
	"fmt"
	"strings"
)

// targeting_bar_colors.go — "creatures you control can't be the targets
// of blue or black spells this turn" (Autumn's Veil, #6263).
//
// Not hexproof from a colour (CR 702.11e refuses only an opponent's
// spells AND abilities) and not protection (CR 702.16b also refuses
// abilities from a source of the quality, and does more than targeting):
// this refuses SPELLS of the named colours, whoever controls them, and no
// ability. So it is its own record: a ScopedEffect with a
// ModCantBeTargetedBySpellsOfColors mod naming the colours, over a live
// scope (ScopeYourCreatures for Autumn's Veil: it changes no
// characteristic, CR 611.2c, so a creature that arrives later this turn
// is covered too). Read by canBeTargetedBy, which both the CR 601.2c
// announce gate and the CR 608.2b resolution re-check go through, so a
// blue spell already aimed at a creature when the Veil resolves loses
// that target as it resolves.

// CantBeTargetedBySpellsOfColorsMod is the mod for the record, with
// `colors` as uppercase WUBRG letters.
func CantBeTargetedBySpellsOfColorsMod(colors ...string) Mod {
	return Mod{Kind: ModCantBeTargetedBySpellsOfColors, Colors: colors}
}

// colorSpellTargetingBarLocked reports whether a live record bars the
// spell `src` from targeting the permanent `c`: some record carries the
// mod over `c` and the spell has one of its colours (read off the
// source's characteristics, so a spell an effect made blue is blue).
// Nil-receiver safe. Caller must hold g.mu. Reads only.
func (g *Game) colorSpellTargetingBarLocked(c *Card, src TargetSource) bool {
	if g == nil || c == nil || !src.IsSpell() || len(g.ScopedEffects) == 0 {
		return false
	}
	var colors []string
	if ch := src.Characteristics(); ch != nil {
		colors = ch.Colors
	}
	if len(colors) == 0 {
		return false
	}
	for i := range g.ScopedEffects {
		e := &g.ScopedEffects[i]
		for _, m := range e.Mods {
			if m.Kind != ModCantBeTargetedBySpellsOfColors || !colorsOverlap(colors, m.Colors) {
				continue
			}
			covers := affectedPredicate(e.Affected)
			if e.Scope != ScopeNone {
				covers = scopePredicate(e.Scope, e.Controller)
			}
			if covers(c, g, nil) {
				return true
			}
		}
	}
	return false
}

// colorsOverlap reports whether two colour lists share a colour.
func colorsOverlap(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if strings.EqualFold(x, y) {
				return true
			}
		}
	}
	return false
}

// colorSpellTargetingBarModProblem is the restore check for the mod: it
// names at least one colour, every one a WUBRG letter. "" when sound.
func colorSpellTargetingBarModProblem(m Mod) string {
	if m.Kind != ModCantBeTargetedBySpellsOfColors {
		return ""
	}
	if len(m.Colors) == 0 {
		return fmt.Sprintf("a %s mod that names no colour", m.Kind)
	}
	for _, c := range m.Colors {
		if len(c) != 1 || !strings.Contains("WUBRG", c) {
			return fmt.Sprintf("a %s mod naming %q, not a colour", m.Kind, c)
		}
	}
	return ""
}

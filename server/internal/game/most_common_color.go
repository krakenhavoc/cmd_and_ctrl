package game

import (
	"strings"

	"github.com/google/uuid"
)

// most_common_color.go — #6207: "the most common color among all
// permanents or tied for most common" (the Mirage Djinns, Heroic
// Defiance, Tsabo's Assassin, Barrin's Unmaking) and "the most common
// color among nontoken permanents the chosen player controls" (Call to
// Arms).
//
// Each permanent counts once for each of its colours (CR 105.2: an
// object can be one or more of the five colours, or no colour at all),
// read through the layered view, so a permanent an effect made white
// counts as white. Colourless is not a colour and counts for nothing.
// A colour is "the most common" when no colour has more permanents and
// at least one permanent has it: with no coloured permanent at all, no
// colour is common.
//
// Inside a layer pass this is asked from layer 7c (a Djinn's -2/-2),
// after layer 5 has settled every permanent's colours for the pass
// (layerPassWithLocked applies each layer to every permanent before the
// next), so the count never depends on the P/T effect reading it.

// MostCommonColorsForEffect returns the colours (uppercase WUBRG
// letters, in WUBRG order) that are the most common among the
// permanents on the battlefield, ties included. A non-nil `controller`
// narrows the count to that player's permanents, and `nontokenOnly`
// to nontoken ones (Call to Arms). Nil for no coloured permanent.
//
// The *ForEffect surface: caller must hold g.mu (read or write).
func (g *Game) MostCommonColorsForEffect(controller uuid.UUID, nontokenOnly bool) []string {
	if g.Battlefield == nil {
		return nil
	}
	counts := map[string]int{}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if controller != uuid.Nil && c.Controller != controller {
			continue
		}
		if nontokenOnly && IsTokenPtr(c) {
			continue
		}
		seen := map[string]bool{}
		for _, col := range c.EffectiveColors() {
			col = strings.ToUpper(col)
			if !seen[col] && strings.Contains("WUBRG", col) && len(col) == 1 {
				seen[col] = true
				counts[col]++
			}
		}
	}
	best := 0
	for _, n := range counts {
		best = max(best, n)
	}
	if best == 0 {
		return nil
	}
	var out []string
	for _, col := range []string{"W", "U", "B", "R", "G"} {
		if counts[col] == best {
			out = append(out, col)
		}
	}
	return out
}

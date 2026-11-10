package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b39_helpers.go — shared pieces of Standard set batch 39–42 cards.

// colorsAmongMonocoloredPermanentsYouControl is "each color among
// monocolored permanents you control": the union of the colors of every
// permanent the controller controls that has exactly one color (CR 105.2a),
// in WUBRG order. A multicolored or colorless permanent contributes nothing.
func colorsAmongMonocoloredPermanentsYouControl(g *game.Game, controller uuid.UUID) []string {
	seen := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller {
			continue
		}
		cols := c.EffectiveColors()
		if len(cols) != 1 {
			continue
		}
		seen[strings.ToUpper(cols[0])] = true
	}
	var out []string
	for _, col := range []string{"W", "U", "B", "R", "G"} {
		if seen[col] {
			out = append(out, col)
		}
	}
	return out
}

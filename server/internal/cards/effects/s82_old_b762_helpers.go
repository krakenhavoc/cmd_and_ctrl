package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b762_helpers.go — shared pieces of the Tempest batch 762 cards
// (#5628).

// S82OldB762ShapeshifterKeywords are the abilities an opponent's creatures not
// named `name` give the Shapeshifter: flying, first strike, trample and
// protection from each colour, once each.
func S82OldB762ShapeshifterKeywords(g *game.Game, source *game.Card, name string) []string {
	if g == nil || g.Battlefield == nil || source == nil {
		return nil
	}
	watch := []string{"flying", "first strike", "trample"}
	for _, color := range []string{"white", "blue", "black", "red", "green"} {
		watch = append(watch, "protection from "+color)
	}
	var out []string
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == source.Controller || !c.IsCreature() || strings.EqualFold(c.Name, name) {
			continue
		}
		for _, kw := range watch {
			if game.HasKeyword(c, kw) && !s82OldB762Contains(out, kw) {
				out = append(out, kw)
			}
		}
	}
	return out
}

func s82OldB762Contains(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

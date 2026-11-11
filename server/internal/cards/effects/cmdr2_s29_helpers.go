package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2S29FourCardTypes is the set rule "with four or more card types among
// them" over a picked set of cards (Winter, Cynical Opportunist): the distinct
// card types on the cards' type lines, counted as b16CardTypesInGraveyard does.
// It is a function of the cards alone and captures nothing, as a registered
// set rule must.
func cmdr2S29FourCardTypes(picked []game.Card) bool {
	seen := map[string]bool{}
	for _, c := range picked {
		_, types, _ := game.ParseTypeLine(c.TypeLine)
		for _, t := range types {
			seen[t] = true
		}
	}
	return len(seen) >= 4
}

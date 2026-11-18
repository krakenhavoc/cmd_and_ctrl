package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_rpr_b26_helpers.go — shared bodies for the reprint pass, batch 26
// (#5908).

// rprB26OnePerOtherRedAndGreenCreature is the count of Ulasht, the Hate
// Seed's "a +1/+1 counter on it for each other red creature you control and
// a +1/+1 counter on it for each other green creature you control", for
// b19EntersWithCountersCounted: one per other red creature plus one per
// other green creature its controller controls as it enters, colours read
// through the layers. A red-and-green creature counts for both clauses, so
// it is two counters. Ulasht is still the stack object, so it is never
// counted itself.
func rprB26OnePerOtherRedAndGreenCreature(g *game.Game, src *game.Card) int {
	red, green := src.TextColor("R"), src.TextColor("G")
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != src.Controller || c.InstanceID == src.InstanceID || !c.IsCreature() {
			continue
		}
		if c.HasColorWord(red) {
			n++
		}
		if c.HasColorWord(green) {
			n++
		}
	}
	return n
}

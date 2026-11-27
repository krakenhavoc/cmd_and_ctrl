package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b601_helpers.go — shared shape of S82 OLD batch 601 (Betrayers of
// Kamigawa, #5602): counting the permanents of a creature type a player
// controls.

// OldB601CountSubtypeWordYouControl is "the number of <word>s you control"
// for `controller`: every permanent they control whose types, read through
// the layers, include the word. A card passes source.TextSubtype("Spirit")
// so a text-changing effect on the source reaches the word (CR 612.2).
func OldB601CountSubtypeWordYouControl(g *game.Game, controller uuid.UUID, word string) int {
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == controller && c.HasSubtypeWord(word) {
			n++
		}
	}
	return n
}

// OldB601SoulshiftX is Kodama of the Center Tree's X, "the number of
// Spirits you control", read for the caster as the soulshift target is
// chosen. The ability has no source in a target predicate, so the word is
// Spirit as printed.
func OldB601SoulshiftX(g *game.Game, caster uuid.UUID) int {
	return OldB601CountSubtypeWordYouControl(g, caster, "Spirit")
}

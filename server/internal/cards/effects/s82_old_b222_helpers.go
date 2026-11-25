package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b222_helpers.go — shared bodies for Ixalan, batch 222 (#5532).

// OldB222YouControlAPlaneswalker is "you control a <word> planeswalker":
// a planeswalker `you` controls with that subtype (Jace's Sentinel).
func OldB222YouControlAPlaneswalker(g *game.Game, you uuid.UUID, word string) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.IsPlaneswalker() && c.HasSubtype(word) {
			return true
		}
	}
	return false
}

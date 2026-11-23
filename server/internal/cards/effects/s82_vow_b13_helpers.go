package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b13_helpers.go — shared by Geistlight Snare and Supernatural
// Rescue (Innistrad: Crimson Vow, S82).

// vowB13YouControlASpirit is "if you control a Spirit" for a timing
// permission read as the cast begins (CR 601.3); the spell is still in
// the hand then, so it is not the Spirit it asks about.
func VowB13YouControlASpirit(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.HasSubtype("Spirit") {
			return true
		}
	}
	return false
}

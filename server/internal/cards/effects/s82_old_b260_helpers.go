package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b260_helpers.go — shared pieces of Kaladesh batch 260 (#5539).

// S82OldB260ArtifactsYouControl is the number of artifacts `you` controls,
// read through the layered view (a permanent made an artifact by another
// effect counts), for "if you control an artifact" and "three or more
// artifacts" (Toolcraft Exemplar).
func S82OldB260ArtifactsYouControl(g *game.Game, you uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.IsArtifact() {
			n++
		}
	}
	return n
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b729_helpers.go — Urza's Saga, batch 729 (#5624, S82).

// OldB729OwnCreatureCandidates is ChoosePermanents.Candidates for "choose a
// creature you control" as one seat is asked: every creature that seat
// controls, exactly one to be picked. Curfew's and Time Wipe's pick.
func OldB729OwnCreatureCandidates(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == of && c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out, 1, 1
}

package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// memory_vessel_test.go — #2559: every player plays the cards they
// exiled, and nobody plays from their hand, until the activator's next
// turn (Memory Vessel); and the same permission without the ban, until
// your next end step (Rocco, Street Chef).

const (
	MemoryVesselOracle = "0179bc62-823e-46b9-b536-342904fedafc"
	RoccoOracle        = "071f5b4b-6a1e-4dc3-9c08-b79713191811"
)

func PermissionFor(g *game.Game, player, card uuid.UUID) *game.CastPermission {
	var perm *game.CastPermission
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		perm = g.CastPermissionForLocked(player, c, game.ZoneExile)
	})
	return perm
}

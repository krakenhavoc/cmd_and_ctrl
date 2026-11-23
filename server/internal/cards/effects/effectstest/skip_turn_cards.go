package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// skip_turn_cards_test.go — "skip your next turn" (CR 614.10, ADR 0188
// Decision 1, #3559): Magosi, the Waterveil and Lethal Vapors. Each card
// is driven through the real activation and the real rotation seam, and
// the skip is asserted by who is active after the turn ends.

const (
	MagosiOracle       = "4bdffa67-e6b3-4588-b76e-c11db6f043ca"
	LethalVaporsOracle = "9fe9650d-6c9b-4988-9ea4-79ba0d55a77b"
)

// --- Magosi, the Waterveil ------------------------------------------

func MagosiOnBattlefield(g *game.Game, owner uuid.UUID) uuid.UUID {
	return PushCatalogPermanent(g, owner, "Magosi, the Waterveil", "Land", MagosiOracle, false)
}

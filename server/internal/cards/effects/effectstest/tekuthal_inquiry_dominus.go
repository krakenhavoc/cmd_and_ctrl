package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tekuthal_inquiry_dominus_test.go — #943's card. The component's own
// rules are pinned in game/counter_cost_any_kind_test.go; this is
// about Tekuthal doing what its oracle text says.

const TekuthalOracle = "4716ab91-30e6-4c63-8389-a9db8f9414d8"

// pushTekuthal seats Tekuthal under `owner`'s control.
func PushTekuthal(g *game.Game, owner uuid.UUID) uuid.UUID {
	return B12Push(g, owner, "Tekuthal, Inquiry Dominus", "Legendary Creature — Phyrexian Horror", TekuthalOracle, 3, 5)
}

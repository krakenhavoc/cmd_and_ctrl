package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// waterbend_cost_cards_test.go — the proof cards for #1310 (waterbend
// as an activated ability's cost) and #1311 (a ward whose cost is a
// waterbend cost). The engine shapes are pinned in
// game/waterbend_cost_test.go; these pin each card's printed text.

const (
	KataraWaterTribesHopeOracle = "234fb291-0b62-4092-9071-81311c71bd53"
	TheUnagiOfKyoshiIslandOracl = "be922360-ee9f-4cfb-bf10-d481c5de5cb1"
)

func IsTapped(g *game.Game, id uuid.UUID) bool {
	c := BattlefieldCardFor(g, id)
	return c != nil && c.Tapped
}

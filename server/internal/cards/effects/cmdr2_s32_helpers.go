package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s32_helpers.go — shared pieces of Commander deck batch 32
// (Foreboding Steamboat, Bronzebeak Foragers).

// cmdr2S32PutExiledCardIntoGraveyard is "put [a card exiled with this] into
// its owner's graveyard": the card moves from exile to its owner's graveyard
// (CR 400.3) if it is still in exile. It reports whether it moved, for the
// "if you do" a clause may print after it. The card leaves exile, so the
// CR 607.2a link and the "until" record that held it are spent: the card is
// not returned when its source leaves.
func cmdr2S32PutExiledCardIntoGraveyard(g *game.Game, id uuid.UUID) (bool, error) {
	z := g.FindCardZoneForEffect(id)
	if z == nil || z.Kind != game.ZoneExile {
		return false, nil
	}
	if err := g.PutIntoGraveyardForEffect(id); err != nil {
		return false, err
	}
	return true, nil
}

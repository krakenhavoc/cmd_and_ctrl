package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// baseOneByOne is "a creature with base power and toughness 1/1" (CR 208.4b):
// the values after layer 7b, so a pump, an anthem or a counter does not
// change it but a copy effect or a "base power and toughness" effect does.
func baseOneByOne(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	return c.IsCreature() && c.BasePower() == 1 && c.BaseToughness() == 1
}

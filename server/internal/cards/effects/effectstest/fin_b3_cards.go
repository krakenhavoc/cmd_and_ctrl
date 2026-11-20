package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// finB3Has reports whether the card, read through the layers, has the
// ability.
func FinB3Has(c game.Card, ability string) bool {
	for _, a := range c.Effective().Abilities {
		if a == ability {
			return true
		}
	}
	return false
}

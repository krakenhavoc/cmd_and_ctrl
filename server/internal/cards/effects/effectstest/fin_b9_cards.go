package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func FinB9FillParty(g *game.Game, owner uuid.UUID, skip string) {
	for _, ty := range []string{"Cleric", "Rogue", "Warrior", "Wizard"} {
		if ty != skip {
			B12Creature(g, owner, "Party "+ty, "Creature — Human "+ty, 1, 1)
		}
	}
}

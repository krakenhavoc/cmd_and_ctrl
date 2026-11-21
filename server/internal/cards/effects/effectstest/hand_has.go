package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hand_has.go — a hand-membership check shared across test packages.

func HandHas(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

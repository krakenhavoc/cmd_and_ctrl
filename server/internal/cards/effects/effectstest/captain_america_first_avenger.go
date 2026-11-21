package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// capGear puts an Equipment with the given mana cost on the battlefield
// under `controller`, attached to `host` (uuid.Nil: attached to nothing).
func CapGear(g *game.Game, controller, host uuid.UUID, name, cost string) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: EquipTypeLine, ManaCost: cost,
		Owner: controller, Controller: controller,
	}
	if host != uuid.Nil {
		c.AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	}
	return PushBattlefieldCardWithTimestamp(g, c)
}

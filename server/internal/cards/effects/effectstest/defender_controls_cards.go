package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func PushTestPermanent(g *game.Game, controller uuid.UUID, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = controller, controller
	return PushBattlefieldCardWithTimestamp(g, c)
}

// declareResult tries one attack on a clone, so each check starts fresh.
func DeclareResult(g *game.Game, attacker, target uuid.UUID) error {
	return g.Clone().DeclareAttacker(attacker, target)
}

package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func Cmdr2B13Activate(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID, params game.ActivateAbilityParams, colours ...string) error {
	t.Helper()
	me.ManaPool.EmptyPool()
	for _, c := range colours {
		FillPoolColored(me, c, 1)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, params); err != nil {
		return err
	}
	PassPriorityAroundTable(t, g)
	return nil
}

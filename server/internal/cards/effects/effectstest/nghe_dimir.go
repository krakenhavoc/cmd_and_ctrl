package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ngheAnimate activates row 0 of source with the given coloured mana
// plus `generic` colourless, and settles the stack.
func NgheAnimate(t *testing.T, g *game.Game, me *game.Player, source uuid.UUID, generic int, colours ...string) {
	t.Helper()
	me.ManaPool.EmptyPool()
	FillPool(me, generic)
	for _, c := range colours {
		FillPoolColored(me, c, 1)
	}
	S58p6Activate(t, g, me.ID, source, game.ActivateAbilityParams{})
	PassPriorityAroundTable(t, g)
}

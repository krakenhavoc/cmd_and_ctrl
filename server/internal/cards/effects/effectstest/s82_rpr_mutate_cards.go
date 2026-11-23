package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rprMutatorInHand puts a catalog mutate card in `owner`'s hand and
// walks to a main phase.
func RprMutatorInHand(t *testing.T, g *game.Game, owner uuid.UUID, c game.Card) uuid.UUID {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = owner, owner
	g.PlayerByID(owner).Hand.PushTop(c)
	AdvanceToMainForTest(t, g)
	return c.InstanceID
}

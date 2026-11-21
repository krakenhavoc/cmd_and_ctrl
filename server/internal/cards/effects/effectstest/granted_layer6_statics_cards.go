package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// loseAllAbilities pins a "loses all abilities" record to a permanent,
// with a timestamp later than everything already on the table.
func Gl6LoseAllAbilities(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(id),
			[]game.Mod{game.LoseAllAbilitiesMod()}, game.IndefiniteDuration(), "test — loses all abilities")
	})
}

func Gl6HasKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	return ContainsString(EffectiveAbilities(t, g, id), kw)
}

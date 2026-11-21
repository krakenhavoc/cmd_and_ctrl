package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func Spm8Live(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := FindBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return *c
}

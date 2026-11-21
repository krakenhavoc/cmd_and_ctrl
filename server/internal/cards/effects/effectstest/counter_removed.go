package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func CrRemove(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	PutCounters(t, g, id, kind, -n)
}

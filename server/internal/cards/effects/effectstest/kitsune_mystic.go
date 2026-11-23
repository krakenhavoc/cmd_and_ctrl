package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func AttachedToNow(t *testing.T, g *game.Game, aura, host uuid.UUID) bool {
	t.Helper()
	c := FindBattlefieldCardForTest(g, aura)
	return c != nil && c.IsAttachedTo(host)
}

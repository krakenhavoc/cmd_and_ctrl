package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- shared test helpers --------------------------------------------------------------

// somB1Equip puts an Equipment on the battlefield and attaches it to host.
func SomB1Equip(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	eq := B12Push(g, owner, name, "Artifact — Equipment", oracle, 0, 0)
	AttachForTest(t, g, eq, host)
	return eq
}

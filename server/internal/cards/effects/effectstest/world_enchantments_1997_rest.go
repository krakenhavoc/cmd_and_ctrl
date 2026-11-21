package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Arboria -------------------------------------------------------

// arbTargets is who the enumerator offers `attacker` at, as ids.
func ArbTargets(t *testing.T, g *game.Game, seat, attacker uuid.UUID) map[string]bool {
	t.Helper()
	return BloodthirsterTargets(t, g, seat, attacker)
}

package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attackInto declares each attacker at its defender, locks the
// declaration in and walks into declare blockers.
func AttackInto(t *testing.T, g *game.Game, attacks map[uuid.UUID]uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	for attacker, defender := range attacks {
		if err := g.DeclareAttacker(attacker, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	LockInAttacks(t, g)
	AdvanceTo(t, g, game.StepDeclareBlockers)
}

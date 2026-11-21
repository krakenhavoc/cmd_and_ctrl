package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- fixtures ------------------------------------------------------------------

// lciB1Attack declares the attacker against defender and locks the
// declaration in, so "whenever this attacks" triggers are harvested, then
// resolves what they put on the stack.
func LciB1Attack(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	LockInAttacks(t, g)
	PassPriorityAroundTable(t, g)
}

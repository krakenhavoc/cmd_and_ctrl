package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// socB34AttackWith declares `attackers` against `defender` from the declare
// attackers step of seat 0 and settles the attack triggers.
func SocB34AttackWith(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	AdvanceToDeclareAttackersOf(t, g, 0)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	LockInAttacks(t, g)
	PassPriorityAroundTable(t, g)
}

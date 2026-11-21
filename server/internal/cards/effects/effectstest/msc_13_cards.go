package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc13BlockRefused declares `attackers` against seat `defender`, advances to
// the declare-blockers step and reports for each attacker whether the
// blocker was refused.
func Msc13BlockRefused(t *testing.T, g *game.Game, defender uuid.UUID, blocker uuid.UUID, attackers ...uuid.UUID) map[uuid.UUID]bool {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	LockInAttacks(t, g)
	PassPriorityAroundTable(t, g)
	AdvanceTo(t, g, game.StepDeclareBlockers)
	out := map[uuid.UUID]bool{}
	for _, a := range attackers {
		out[a] = g.DeclareBlocker(blocker, a) != nil
	}
	return out
}

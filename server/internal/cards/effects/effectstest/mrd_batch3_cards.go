package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mrd3BlockedBy declares `attacker` attacking `defender` and blocked by
// `blocker`, with the blocks locked in.
func Mrd3BlockedBy(t *testing.T, g *game.Game, defender, attacker, blocker uuid.UUID) {
	t.Helper()
	DeclareAttack(t, g, defender, attacker)
	AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	LockInBlocks(t, g)
}

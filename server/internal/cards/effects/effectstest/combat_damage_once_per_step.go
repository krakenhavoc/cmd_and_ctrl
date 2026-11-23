package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// onceStepBlock attacks the defender with the attacker, blocks it with
// the blockers, and stops at the first combat damage step.
func OnceStepBlock(t *testing.T, g *game.Game, defender, attacker uuid.UUID, step game.Step, blockers ...uuid.UUID) {
	t.Helper()
	DeclareAttack(t, g, defender, attacker)
	AdvanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range blockers {
		if err := g.DeclareBlocker(b, attacker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	LockInBlocks(t, g)
	AdvanceTo(t, g, step)
}

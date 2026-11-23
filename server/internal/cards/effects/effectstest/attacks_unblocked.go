package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attacks_unblocked_test.go — #1279's proof cards. "Whenever this
// creature attacks and isn't blocked" (CR 509.3) fires off the
// defending player's COMPLETED block declaration, so it fires when the
// defender finishes declaring — not when the step begins, and never
// while they are still deciding.

const (
	SwampMosquitoOracle        = "4bb4844c-2678-45c7-8f5e-9cf185fd484b"
	AbyssalNightstalkerOracle  = "10733767-1c97-4e2e-b02b-54bf346f6583"
	GuiltfeederOracle          = "0285bd20-f49e-48f7-8c8c-960f9fbe4d34"
	EternalOfHarshTruthsOracle = "5c09d646-afea-4fd0-9752-04820b81cc5b"
)

// unblockedAttackIntoBlockers declares `attacker` at seat 1 and walks
// into declare_blockers.
func UnblockedAttackIntoBlockers(t *testing.T, g *game.Game, attacker uuid.UUID) {
	t.Helper()
	DeclareAttack(t, g, g.Seats[1].ID, attacker)
	AdvanceTo(t, g, game.StepDeclareBlockers)
}

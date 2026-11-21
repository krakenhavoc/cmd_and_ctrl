package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trekB25ToCombatDamage declares attackers at defender and walks into the
// combat damage step, settling what fires on the way.
func TrekB25ToCombatDamage(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	TrekB18AAttack(t, g, defender, attackers...)
	AdvanceTo(t, g, game.StepCombatDamage)
	PassPriorityAroundTable(t, g)
}

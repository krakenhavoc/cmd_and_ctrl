package effectstest

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// faMustAttack reports whether the active player cannot pass in the
// declare attackers step without attacking.
func FaMustAttack(t *testing.T, g *game.Game) bool {
	t.Helper()
	AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	err := g.PassPriority()
	return errors.Is(err, game.ErrAttackRequirement)
}

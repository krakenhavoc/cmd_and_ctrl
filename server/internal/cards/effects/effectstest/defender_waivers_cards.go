package effectstest

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wantDefenderRefused asserts the attack is refused for defender.
func WantDefenderRefused(t *testing.T, g *game.Game, attacker, target uuid.UUID) {
	t.Helper()
	if err := DeclareResult(g, attacker, target); !errors.Is(err, game.ErrDefender) {
		t.Errorf("attack %s: err = %v, want ErrDefender", target, err)
	}
}

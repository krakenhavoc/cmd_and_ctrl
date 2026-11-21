package effectstest

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func WantAttackRefused(t *testing.T, g *game.Game, attacker, target uuid.UUID, why string) {
	t.Helper()
	err := DeclareResult(g, attacker, target)
	if !errors.Is(err, game.ErrIllegalAttackTarget) {
		t.Fatalf("attack %s: err = %v, want ErrIllegalAttackTarget", target, err)
	}
	if why != "" && !strings.Contains(err.Error(), why) {
		t.Errorf("refusal %q does not say %q", err.Error(), why)
	}
}

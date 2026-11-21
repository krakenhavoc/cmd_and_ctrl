package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// turn_purpose_test.go — ADR 0188 Decision 10, #3566: the catalog cards
// that give or cost whole turns declare how many, so the heuristic can
// price a turn (Weights.Turn), and Register refuses a declaration that
// would price the wrong player's turn.

func SpecNamed(t *testing.T, name string) effects.Spec {
	t.Helper()
	for _, s := range effects.All() {
		if s.Name == name {
			return s
		}
	}
	t.Fatalf("no catalog card named %q", name)
	return effects.Spec{}
}

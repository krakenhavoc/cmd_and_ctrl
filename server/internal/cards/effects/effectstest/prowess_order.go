package effectstest

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prowess_order_test.go — #1511. A batch of prowess triggers from
// different creatures commutes, so the CR 603.3b ordering prompt is
// skipped for it (ADR 0018's #1511 amendment). This skip reads no
// board; a batch with anything else in it — even one untargeted
// trigger of another kind — is left to the #2884 independence check,
// which reads prowess's declared footprint since #3013. The targeted
// case is pinned at the drain in game/trigger_order_commutes_test.go
// and, since #1529 held the drain for a trigger that is still choosing
// its target, end to end from a cast in trigger_batch_order_test.go
// (prowess plus Caldera Pyremaw).

// settleWithoutPrompts passes priority until the stack settles and
// fails on ANY pending choice — the whole point of the auto-ordered
// shape is that the controller is never asked anything.
func SettleWithoutPrompts(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if len(g.PendingChoices) > 0 {
			t.Fatalf("a prompt opened while an all-prowess batch settled: %+v", g.PendingChoices[0])
		}
		if StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				continue
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}

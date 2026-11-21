package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lki_copy_test.go — #1255, CR 608.2h against CR 608.2b.
//
// A copy effect that NAMES a spell without targeting it copies the
// spell from last-known information once it has left the stack; one
// that TARGETS the spell does nothing, because its only target is
// illegal. Storm is pinned in storm_test.go. This file holds the other
// two non-targeting families that opted in (Thousand-Year Storm and
// Doublecast's "copy that spell") and the targeting family that must
// NOT have changed (Reverberate).

// counterNow counters a spell on the stack, as a counterspell cast in
// response would.
func CounterNow(t *testing.T, g *game.Game, spell uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(spell); err != nil {
			t.Fatalf("CounterTargetForEffect: %v", err)
		}
	})
	if g.Stack.Contains(spell) {
		t.Fatal("the spell was not countered — the fixture is wrong")
	}
}

// settleKeepingTarget passes priority until the stack is empty,
// answering every copy's re-target prompt with `victim` and every
// trigger-order prompt as offered. Returns how many re-target prompts
// it answered.
func SettleKeepingTarget(t *testing.T, g *game.Game, chooser, victim uuid.UUID) int {
	t.Helper()
	answered := 0
	for i := 0; i < 64; i++ {
		if p := TriggerOrderPromptFor(g, chooser); p != nil {
			if err := g.ResolveTriggerOrder(p.ID, chooser, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := LatestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser,
				game.TargetRef{Kind: game.TargetPlayer, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			answered++
			continue
		}
		if StackFullyEmpty(g) {
			return answered
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
	return answered
}

package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ringTempt has the Ring tempt `player` directly, as a resolving effect
// would, and reports what the continuation was handed (uuid.Nil until
// it runs — it does not run while a ring_bearer prompt is open).
func RingTempt(t *testing.T, g *game.Game, player uuid.UUID) *uuid.UUID {
	t.Helper()
	got := new(uuid.UUID)
	*got = uuid.Max
	g.WithWriteLock(func() {
		if err := g.RingTemptsKeyedForEffect(player, uuid.Nil, HookTail(func(_ *game.Game, o game.TailOutcome) error {
			rb := tailFirstID(o)
			*got = rb
			return nil
		})); err != nil {
			t.Fatalf("RingTemptsKeyedForEffect: %v", err)
		}
	})
	return got
}

// ringPrompt is the open ring_bearer prompt `player` owes, or nil.
func RingPrompt(g *game.Game, player uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceRingBearer && c.Chooser == player {
			return c
		}
	}
	return nil
}

// answerRing answers `player`'s open ring_bearer prompt with `pick`.
func AnswerRing(t *testing.T, g *game.Game, player, pick uuid.UUID) {
	t.Helper()
	c := RingPrompt(g, player)
	if c == nil {
		t.Fatal("no ring_bearer prompt is open")
	}
	if err := g.ResolveRingBearer(c.ID, player, []uuid.UUID{pick}); err != nil {
		t.Fatalf("ResolveRingBearer: %v", err)
	}
}

// ringBearerOf reads `player`'s Ring-bearer under the lock.
func RingBearerOf(g *game.Game, player uuid.UUID) uuid.UUID {
	var id uuid.UUID
	g.ReadSnapshot(func() { id = game.RingBearerOf(g, player) })
	return id
}

func RingCount(g *game.Game, player uuid.UUID) int {
	var n int
	g.ReadSnapshot(func() { n = game.RingTemptCount(g, player) })
	return n
}

func RingBFCard(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	t.Fatalf("%s is not on the battlefield", id)
	return nil
}

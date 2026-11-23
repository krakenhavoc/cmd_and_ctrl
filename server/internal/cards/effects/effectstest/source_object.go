package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// source_object_test.go is the card-level proof for #1418: a triggered
// ability names the OBJECT its source was when it triggered, so "this
// permanent" at resolution is never a new object the same card has
// become (CR 400.7). The engine half is game/source_object_test.go.
//
// Every test here returns the source to the battlefield with a raw move
// after bouncing it, so it is the same card under the same instance ID,
// a new object, and — because the raw move emits nothing — without a
// second copy of its own enter trigger to muddy the count.

// flickerInResponse bounces `id` to its owner's hand and puts it
// straight back onto the battlefield while its trigger waits.
func FlickerInResponse(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			t.Fatalf("flicker: %v is nowhere", id)
		}
		owner := g.PlayerByIDForEffect(c.Owner)
		before := c.ObjectEpoch
		if err := g.BounceToHandForEffect(id); err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
		if _, err := game.MoveCard(owner.Hand, g.Battlefield, id); err != nil {
			t.Fatalf("MoveCard back: %v", err)
		}
		g.BumpLayerVersionForTest()
		g.RecomputeLayersIfStaleLocked()
		if back, _ := g.LookupCardForEffect(id); back.ObjectEpoch == before {
			t.Fatal("setup: the returned card should be a new object")
		}
	})
}

package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyword_reading_grant_test.go — #2663, ADR 0067 amendment 2026-10-23:
// a layer-6 grant that reads a keyword ("Other creatures you control with
// flying have indestructible") waits for every effect that adds or removes
// that keyword (CR 613.8a, 613.8b), against the real catalog.

const (
	SepharaOracle     = "ac8952c9-bd30-48cd-b964-7e6271f9cbc8"
	KwendeOracle      = "3adebbbf-39fe-4b42-b719-2f8edb009693"
	AetherChaserOracl = "8aab0b55-2779-43c7-a3ee-151b8e5c72f3"
)

// krgHas forces a recompute and asks whether the battlefield card has the
// keyword.
func KrgHas(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	var has, found bool
	g.ReadSnapshot(func() {
		c, ok := BattlefieldCard(g, id)
		found = ok
		has = ok && game.HasKeyword(&c, kw)
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return has
}

// krgGrantUntilEOT gives the creature keywords until end of turn, as a
// resolving spell would (Jump), with a timestamp later than everything
// already on the battlefield.
func KrgGrantUntilEOT(t *testing.T, g *game.Game, controller, target uuid.UUID, kws ...string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := (effects.GrantKeywordUntilEOT{Target: target, Keywords: kws}).Apply(effects.NewContext(g, &game.StackItem{Controller: controller})); err != nil {
			t.Fatal(err)
		}
	})
}

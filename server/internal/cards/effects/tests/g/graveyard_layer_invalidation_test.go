package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyard_layer_invalidation_test.go is #1117's graveyard half from
// the catalog side: a real Lhurgoyf, a real mill, and no explicit
// BumpLayerVersionForTest anywhere.
//
// The existing Tarmogoyf tests all seed graveyards through
// pushGraveyardCardWithTypeLine, which bumps the layer version by
// hand — correct for a raw push, which fires no event, but it means
// none of them could see that a REAL mill did not invalidate either.
// This one drives the engine's own mill path, so it fails on the
// pre-#1117 listener.

// TestTarmogoyfUpdatesAfterARealMill mills two typed cards off the
// top of a library and reads the CDA with nothing in between.
func TestTarmogoyfUpdatesAfterARealMill(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0]
	goyf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Tarmogoyf",
		TypeLine:   "Creature — Lhurgoyf",
		OracleID:   effectstest.TarmogoyfOracle,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	if got := effectstest.EffectivePower(t, g, goyf); got != 0 {
		t.Fatalf("power with empty graveyards = %d, want 0 (sanity)", got)
	}

	// The stock catalog deck is untyped filler, so put cards with
	// real card types where the mill will find them.
	g.WithWriteLock(func() {
		owner.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Owner: owner.ID})
		owner.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Lightning Bolt", TypeLine: "Instant", Owner: owner.ID})
	})

	g.WithWriteLock(func() {
		if err := g.MillNForEffect(owner.ID, 2); err != nil {
			t.Fatalf("MillNForEffect: %v", err)
		}
	})

	if got := effectstest.EffectivePower(t, g, goyf); got != 2 {
		t.Errorf("power right after milling a Creature and an Instant = %d, want 2", got)
	}
	if got := effectstest.EffectiveToughness(t, g, goyf); got != 3 {
		t.Errorf("toughness right after the mill = %d, want 3", got)
	}
}

// TestTarmogoyfUpdatesAfterADiscard is the other arrival route — a
// hand → graveyard move, which is an EventDiscardCard rather than an
// EventZoneMove, and is why the bump is keyed on the ZONES an event
// names rather than on its kind.
func TestTarmogoyfUpdatesAfterADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0]
	goyf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Tarmogoyf",
		TypeLine:   "Creature — Lhurgoyf",
		OracleID:   effectstest.TarmogoyfOracle,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	if got := effectstest.EffectivePower(t, g, goyf); got != 0 {
		t.Fatalf("power with empty graveyards = %d, want 0 (sanity)", got)
	}

	// One card in hand, so the random discard can only pick it.
	g.WithWriteLock(func() {
		owner.Hand.Cards = nil
		owner.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Pacifism", TypeLine: "Enchantment — Aura", Owner: owner.ID})
	})
	g.WithWriteLock(func() {
		if err := g.DiscardRandomForEffect(owner.ID, 1); err != nil {
			t.Fatalf("DiscardRandomForEffect: %v", err)
		}
	})

	if got := effectstest.EffectivePower(t, g, goyf); got != 1 {
		t.Errorf("power right after discarding an Enchantment = %d, want 1", got)
	}
}

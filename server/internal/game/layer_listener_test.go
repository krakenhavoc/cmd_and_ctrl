package game

import (
	"testing"

	"github.com/google/uuid"
)

// layer_listener_test.go covers the S16 sub-PR 2 built-in
// layer-version-bump listener. The listener is auto-registered by
// NewGame; tests fire events directly via EmitEvent (under a write
// lock) and assert that g.layerVersion ticks for the events that
// affect continuous-effect resolution and stays put for the ones
// that don't.

// readLayerVersion reads the atomic layer-version counter. Tests use
// this rather than g.layerVersion.Load() directly so the cleaner
// intent is visible at the call site.
func readLayerVersion(g *Game) uint64 {
	return g.layerVersion.Load()
}

// TestLayerVersionBumpsOnBattlefieldEntry proves the listener fires
// when a card moves INTO the battlefield. EventZoneMove with
// NewZone == battlefield is the universal "entered the battlefield"
// signal regardless of whether the call site also emitted EventETB.
func TestLayerVersionBumpsOnBattlefieldEntry(t *testing.T) {
	g := newActiveGame(t)
	before := readLayerVersion(g)

	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			CardID:  uuid.New(),
			OldZone: ZoneHand,
			NewZone: ZoneBattlefield,
		})
	})

	if got := readLayerVersion(g); got <= before {
		t.Errorf("layerVersion did not bump on battlefield entry: was %d, now %d", before, got)
	}
}

// TestLayerVersionBumpsOnBattlefieldExit proves the listener fires
// when a card moves OUT of the battlefield. Symmetric to the entry
// test; covers LTB without coupling to EventLTB emission specifics.
func TestLayerVersionBumpsOnBattlefieldExit(t *testing.T) {
	g := newActiveGame(t)
	before := readLayerVersion(g)

	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			CardID:  uuid.New(),
			OldZone: ZoneBattlefield,
			NewZone: ZoneGraveyard,
		})
	})

	if got := readLayerVersion(g); got <= before {
		t.Errorf("layerVersion did not bump on battlefield exit: was %d, now %d", before, got)
	}
}

// TestLayerVersionBumpsOnCounterChange proves counter mutations
// invalidate the layer cache — Tarmogoyf-style CDA inputs and
// layer 7d's CurrentPower/CurrentToughness delegation both depend
// on counter state.
func TestLayerVersionBumpsOnCounterChange(t *testing.T) {
	g := newActiveGame(t)
	before := readLayerVersion(g)

	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:   EventCounterPlaced,
			CardID: uuid.New(),
			Label:  "+1/+1",
			Amount: 1,
		})
	})

	if got := readLayerVersion(g); got <= before {
		t.Errorf("layerVersion did not bump on counter change: was %d, now %d", before, got)
	}
}

// TestLayerVersionDoesNotBumpOnZoneMoveOutsideBattlefield proves the
// listener is selective: a hand → graveyard discard, library → hand
// draw, etc. should NOT invalidate the layer cache because no static
// ability lives on a non-battlefield permanent (CR 113.6 — battlefield
// is the only zone with continuous effects in S16 scope).
func TestLayerVersionDoesNotBumpOnZoneMoveOutsideBattlefield(t *testing.T) {
	g := newActiveGame(t)
	before := readLayerVersion(g)

	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			CardID:  uuid.New(),
			OldZone: ZoneLibrary,
			NewZone: ZoneHand,
		})
	})

	if got := readLayerVersion(g); got != before {
		t.Errorf("layerVersion bumped on library→hand move: was %d, now %d", before, got)
	}
}

// TestLayerVersionDoesNotBumpOnIrrelevantEvent proves the listener
// is a no-op for unrelated event kinds (life change, draw, cast,
// etc.) — only the events that affect static-ability evaluation
// invalidate the cache.
func TestLayerVersionDoesNotBumpOnIrrelevantEvent(t *testing.T) {
	g := newActiveGame(t)
	before := readLayerVersion(g)

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventDrawCard, CardID: uuid.New()})
		g.EmitEvent(Event{Kind: EventChangeLife, Amount: -3})
		g.EmitEvent(Event{Kind: EventCast, CardID: uuid.New()})
		g.EmitEvent(Event{Kind: EventManaAdded, Source: uuid.New()})
	})

	if got := readLayerVersion(g); got != before {
		t.Errorf("layerVersion bumped on irrelevant events: was %d, now %d", before, got)
	}
}

// TestLayerVersionBumpDrivesRecompute proves the end-to-end loop:
// a battlefield-entry event bumps layerVersion → next ReadSnapshot
// triggers exactly one recompute pass (the fast-path test from
// sub-PR 1 confirmed the no-bump baseline; this confirms the
// bump-then-recompute path).
func TestLayerVersionBumpDrivesRecompute(t *testing.T) {
	g := newActiveGame(t)
	g.ReadSnapshot(func() {}) // settle baseline
	baselineRecomputes := g.LayerRecomputeCountForTest()

	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			CardID:  uuid.New(),
			OldZone: ZoneHand,
			NewZone: ZoneBattlefield,
		})
	})

	g.ReadSnapshot(func() {})
	g.ReadSnapshot(func() {}) // second read should hit the fast path

	got := g.LayerRecomputeCountForTest()
	if got != baselineRecomputes+1 {
		t.Errorf("recompute count = %d, want %d (one bump + N reads = one recompute)", got, baselineRecomputes+1)
	}
}

// TestNewGameRegistersLayerListener proves the auto-registration:
// a freshly-constructed game has the layer-version-bump listener in
// its slice. Regression guard against accidentally removing the
// NewGame append.
func TestNewGameRegistersLayerListener(t *testing.T) {
	g := NewGame()
	if len(g.Listeners) == 0 {
		t.Fatalf("NewGame did not register any listeners")
	}
	found := false
	for _, l := range g.Listeners {
		if _, ok := l.(layerVersionBump); ok {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("NewGame did not register layerVersionBump listener; got %d listeners of types %T...", len(g.Listeners), g.Listeners[0])
	}
}

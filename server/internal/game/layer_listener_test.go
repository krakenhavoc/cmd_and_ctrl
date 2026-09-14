package game

import (
	"fmt"
	"math/rand/v2"
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

// --- the conditional hand bump (#74) -------------------------------
//
// A hand change is the most frequent event in the game, so the
// listener does not invalidate on it unconditionally. It asks the
// narrower question: is anything on the battlefield READING a hand
// right now? Psychosis Crawler is the card; StaticAbility.
// DependsOnHandSize is how it says so.
//
// The pair of tests below is the contract, and the two guards above
// (TestLayerVersionDoesNotBumpOnZoneMoveOutsideBattlefield,
// TestLayerVersionDoesNotBumpOnIrrelevantEvent) are the other half:
// they still pass unchanged, because with no such permanent in play
// this is exactly the no-op they assert.

const handSizeCDAOracle = "hand-size-cda"

// handSizeCDAForTest is a stub static in the Psychosis Crawler shape:
// a layer-7a CDA that declares the hand dependency.
func handSizeCDAForTest() StaticAbility {
	return StaticAbility{
		Layer:             Layer7PT,
		SubLayer:          SubLayer7A_CDA,
		DependsOnHandSize: true,
		AppliesTo: func(target *Card, _ *Game, source *Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *Characteristic, _ *Card, g *Game, source *Card) {
			n := 0
			if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Hand != nil {
				n = p.Hand.Size()
			}
			c.Power, c.Toughness = n, n
		},
	}
}

// TestLayerVersionBumpsOnHandMoveWhenAHandSizeCDAIsLive is the fix.
func TestLayerVersionBumpsOnHandMoveWhenAHandSizeCDAIsLive(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID == handSizeCDAOracle {
			return []StaticAbility{handSizeCDAForTest()}
		}
		return nil
	})
	g := newActiveGame(t)
	owner := g.Seats[0]
	pushTypedTestCard(g, Card{
		Name:       "Hand Size Horror",
		TypeLine:   "Artifact Creature — Horror",
		OracleID:   handSizeCDAOracle,
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	// A real draw is EventDrawCard, NOT EventZoneMove — which is the
	// detail that made the first draft of this fix silently do
	// nothing. The gate reads the zones, not the kind.
	before := readLayerVersion(g)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:    EventDrawCard,
			Actor:   owner.ID,
			CardID:  uuid.New(),
			OldZone: ZoneLibrary,
			NewZone: ZoneHand,
		})
	})
	if got := readLayerVersion(g); got <= before {
		t.Errorf("layerVersion did not bump on a draw while a hand-size CDA was on the battlefield: was %d, now %d", before, got)
	}

	// And the mirror image: a card leaving a hand.
	before = readLayerVersion(g)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:    EventDiscardCard,
			Actor:   owner.ID,
			CardID:  uuid.New(),
			OldZone: ZoneHand,
			NewZone: ZoneGraveyard,
		})
	})
	if got := readLayerVersion(g); got <= before {
		t.Errorf("layerVersion did not bump on a discard while a hand-size CDA was on the battlefield: was %d, now %d", before, got)
	}

	// A plain zone move into hand ("put it into your hand") counts
	// too — Dark Confidant's upkeep flip is not a draw.
	before = readLayerVersion(g)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			CardID:  uuid.New(),
			OldZone: ZoneLibrary,
			NewZone: ZoneHand,
		})
	})
	if got := readLayerVersion(g); got <= before {
		t.Errorf("layerVersion did not bump on a library→hand move: was %d, now %d", before, got)
	}
}

// TestLayerVersionIgnoresAHandMoveForAStaticThatDoesNotDeclareIt is
// the other side of the gate. A permanent whose statics do NOT
// declare the dependency must not turn every draw at the table into a
// recompute — otherwise the flag is decorative and the cost is back.
func TestLayerVersionIgnoresAHandMoveForAStaticThatDoesNotDeclareIt(t *testing.T) {
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		if oracleID != handSizeCDAOracle {
			return nil
		}
		ab := handSizeCDAForTest()
		ab.DependsOnHandSize = false
		return []StaticAbility{ab}
	})
	g := newActiveGame(t)
	owner := g.Seats[0]
	pushTypedTestCard(g, Card{
		Name:       "Ordinary Anthem",
		TypeLine:   "Enchantment",
		OracleID:   handSizeCDAOracle,
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	before := readLayerVersion(g)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{
			Kind:    EventDrawCard,
			Actor:   owner.ID,
			CardID:  uuid.New(),
			OldZone: ZoneLibrary,
			NewZone: ZoneHand,
		})
	})
	if got := readLayerVersion(g); got != before {
		t.Errorf("layerVersion bumped on a draw with no hand-size CDA in play: was %d, now %d", before, got)
	}
}

// BenchmarkHandMoveWithNoHandSizeCDA measures what a draw costs the
// listener at a full four-player board holding nothing that reads a
// hand — the case every table is in almost all of the time, and the
// one the conditional exists to keep cheap.
func BenchmarkHandMoveWithNoHandSizeCDA(b *testing.B) {
	g := benchGameWithBoard(b, 40)
	ev := Event{Kind: EventDrawCard, CardID: uuid.New(), OldZone: ZoneLibrary, NewZone: ZoneHand}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.WithWriteLock(func() { g.EmitEvent(ev) })
	}
}

// benchGameWithBoard builds a started game with n plain permanents on
// the battlefield, none of them carrying a catalog entry.
func benchGameWithBoard(b *testing.B, n int) *Game {
	b.Helper()
	g := NewGame()
	for i := 0; i < 4; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), buildTestDeck(fmt.Sprintf("Commander %d", i+1))); err != nil {
			b.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		b.Fatalf("Start: %v", err)
	}
	owner := g.Seats[0].ID
	for i := 0; i < n; i++ {
		g.Battlefield.PushTop(Card{
			InstanceID: uuid.New(),
			Name:       "Filler",
			TypeLine:   "Creature — Bear",
			Owner:      owner,
			Controller: owner,
		})
	}
	return g
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

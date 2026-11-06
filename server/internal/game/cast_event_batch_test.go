package game

import (
	"testing"

	"github.com/google/uuid"
)

// cast_event_batch_test.go — #4943: a spell cast or a land played as a
// player's own action is its own event (CR 601.2a, 116.2a), so it opens
// its own event batch, as a special action does (#1341). Two casts in one
// priority window are two occurrences for a "one or more" trigger
// (CR 603.2c). Cards/effects' left_graveyard_test.go is the card-facing
// proof (Tormod, the Desecrator and the rest); this is the boundary.

// castEventBatch is the batch of the EventCast for spell id.
func castEventBatch(t *testing.T, g *Game, id uuid.UUID) uint64 {
	t.Helper()
	for i := len(g.Events) - 1; i >= 0; i-- {
		if e := g.Events[i]; e.Kind == EventCast && e.CardID == id {
			return e.Batch
		}
	}
	t.Fatalf("no cast event for %v", id)
	return 0
}

func currentBatchForTest(g *Game) uint64 {
	var b uint64
	g.WithWriteLock(func() { b = g.currentEventBatchLocked() })
	return b
}

func TestTwoCastsInOneWindowAreTwoEventBatches(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	emitProbeEvent(g)
	before := lastEventBatch(g)

	first := pushTypedCardToHand(p, "Lightning Bolt", "Instant")
	if err := g.CastSpell(p.ID, first, CastSpellParams{}); err != nil {
		t.Fatalf("cast the first instant: %v", err)
	}
	b1 := castEventBatch(t, g, first)
	if b1 <= before {
		t.Errorf("the first cast is in batch %d, want greater than %d: a cast opens its own batch", b1, before)
	}
	// Nothing resolves and no step changes between the two casts.
	second := pushTypedCardToHand(p, "Shock", "Instant")
	if err := g.CastSpell(p.ID, second, CastSpellParams{}); err != nil {
		t.Fatalf("cast the second instant: %v", err)
	}
	if b2 := castEventBatch(t, g, second); b2 <= b1 {
		t.Errorf("the second cast is in batch %d, want greater than %d: two casts in one priority window are two occurrences (CR 603.2c)", b2, b1)
	}
}

func TestALandPlayOpensItsOwnEventBatch(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	emitProbeEvent(g)
	before := lastEventBatch(g)

	land := pushTypedCardToHand(p, "Forest", "Basic Land — Forest")
	if err := g.CastSpell(p.ID, land, CastSpellParams{}); err != nil {
		t.Fatalf("play a land: %v", err)
	}
	var played uint64
	for i := len(g.Events) - 1; i >= 0; i-- {
		if e := g.Events[i]; e.Kind == EventETB && e.CardID == land {
			played = e.Batch
			break
		}
	}
	if played <= before {
		t.Errorf("the land play is in batch %d, want greater than %d: a land play is its own event (CR 116.2a)", played, before)
	}
}

// A refused cast happened to nothing, so it opens no batch: a seat
// retrying a cast does not split the window's sandbox moves.
func TestRefusedCastOpensNoEventBatch(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	before := currentBatchForTest(g)
	// The cursor is on the upkeep, so a sorcery is refused.
	sorcery := pushTypedCardToHand(p, "Wrath of God", "Sorcery")
	if err := g.CastSpell(p.ID, sorcery, CastSpellParams{}); err != ErrSorcerySpeedRequired {
		t.Fatalf("sorcery on the upkeep: got %v, want ErrSorcerySpeedRequired", err)
	}
	if after := currentBatchForTest(g); after != before {
		t.Errorf("batch %d after a refused cast, want unchanged %d", after, before)
	}
}

// A cast made while a resolution is open (CR 608.2g: a copy the
// resolution offers, Word of Command's play) belongs to that resolution's
// occurrence: everything one resolution emits is one batch, and the
// resolving slot the boundary would clear is still being read.
func TestACastInsideAnOpenResolutionOpensNoEventBatch(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	before := currentBatchForTest(g)
	g.WithWriteLock(func() { g.resolutionOpen = true })
	id := pushTypedCardToHand(p, "Lightning Bolt", "Instant")
	err := g.CastSpell(p.ID, id, CastSpellParams{})
	g.WithWriteLock(func() { g.resolutionOpen = false })
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if got := castEventBatch(t, g, id); got != before {
		t.Errorf("a cast inside an open resolution is in batch %d, want the resolution's batch %d", got, before)
	}
}

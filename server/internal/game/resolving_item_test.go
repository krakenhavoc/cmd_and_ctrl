package game

import (
	"testing"

	"github.com/google/uuid"
)

// resolving_item_test.go — #920. The slot that keeps a resolving
// spell's stack metadata reachable until the resolution is genuinely
// over, tested as a shape: it is set while the effect runs, it
// survives a pause, the next occurrence clears it, and it never
// widens any other lookup.

// resolvingProbeOracle is the stub catalog key the probe spell below
// resolves under.
const resolvingProbeOracle = "00000000-0000-4000-8000-0000000d0920"

// pushResolvingSpell puts a spell on the stack with its meta and
// stubs the catalog resolver so `effect` is what its OnResolve does.
// A hand-built item rather than a cast, so the test drives
// resolveTopOfStackLocked directly.
func pushResolvingSpell(t *testing.T, g *Game, controller uuid.UUID, effect func(*Game, *StackItem) error) uuid.UUID {
	t.Helper()
	withEffectHooks(t,
		func(g *Game, item *StackItem, key string) error {
			if key != resolvingProbeOracle {
				return nil
			}
			return effect(g, item)
		},
		nil,
		func(key string) bool { return key == resolvingProbeOracle },
	)
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(Card{
			InstanceID: id, Name: "Resolving Probe", TypeLine: "Instant",
			OracleID: resolvingProbeOracle,
			Owner:    controller, Controller: controller,
		})
		if g.StackMeta == nil {
			g.StackMeta = make(map[uuid.UUID]*StackItem)
		}
		g.StackMeta[id] = &StackItem{
			ID: id, Kind: StackItemSpell, Controller: controller, Owner: controller,
			SourceCardID: id, Seq: 1,
		}
	})
	return id
}

// TestResolvingSpellIsFindableFromInsideItsOwnResolution is the bug:
// the meta is out of StackMeta before OnResolve runs, so a spell
// copying itself found nothing.
func TestResolvingSpellIsFindableFromInsideItsOwnResolution(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var foundCard Card
	var foundItem *StackItem
	var inMeta bool
	var id uuid.UUID
	id = pushResolvingSpell(t, g, me.ID, func(g *Game, item *StackItem) error {
		_, inMeta = g.StackMeta[item.ID]
		foundCard, foundItem, _ = g.stackSpellLocked(id)
		return nil
	})
	g.WithWriteLock(func() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			t.Fatalf("resolveTopOfStackLocked: %v", err)
		}
	})
	if inMeta {
		t.Error("the item is still removed from StackMeta before the effect runs")
	}
	if foundItem == nil || foundItem.ID != id {
		t.Fatalf("stackSpellLocked found %v inside the resolution, want the resolving item", foundItem)
	}
	if foundCard.InstanceID != id {
		t.Errorf("the card came back as %s, want %s", foundCard.InstanceID, id)
	}
}

// TestResolvingSlotSurvivesAPauseAndIsClearedByTheNextOccurrence —
// the terminal outcome. The copy decision is a prompt, so the slot has
// to outlive the resolution FUNCTION; the next resolution or step
// entry is where it stops.
func TestResolvingSlotSurvivesAPauseAndIsClearedByTheNextOccurrence(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var id uuid.UUID
	id = pushResolvingSpell(t, g, me.ID, func(g *Game, item *StackItem) error {
		// A paused continuation: the rest of the effect is behind a
		// prompt and runs long after this returns.
		g.QueueConfirmForEffect(ConfirmPrompt{
			Chooser:  me.ID,
			Source:   item.SourceCardID,
			Question: "test — pause here",
			OnAccept: func(g *Game) error {
				if _, _, ok := g.stackSpellLocked(id); !ok {
					t.Error("the resolving item is gone by the time the prompt is answered")
				}
				return nil
			},
		})
		return nil
	})
	g.WithWriteLock(func() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			t.Fatalf("resolveTopOfStackLocked: %v", err)
		}
	})
	// The spell itself has already been routed to the graveyard while
	// the question is still open — which is exactly why the card has
	// to be held by value.
	if !me.Graveyard.Contains(id) {
		t.Fatalf("the spell should already be in the graveyard: %+v", me.Graveyard.Cards)
	}
	c := latestChoiceOfKindForTest(g, PendingChoiceConfirm)
	if c == nil {
		t.Fatalf("the prompt is open: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}

	// The next occurrence is the terminal boundary.
	g.WithWriteLock(func() { g.beginEventBatchLocked() })
	if _, _, ok := g.WithWriteLockSpellLookup(id); ok {
		t.Error("the slot outlived its own occurrence")
	}
}

// WithWriteLockSpellLookup is a test-only wrapper so the assertion
// above reads in one line.
func (g *Game) WithWriteLockSpellLookup(id uuid.UUID) (Card, *StackItem, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.stackSpellLocked(id)
}

// TestResolvingSlotDoesNotWidenAnyOtherLookup — a spell that is not
// the resolving one is still not found, so Reverberate pointed at a
// spell countered in response still gets ErrCardNotFound.
func TestResolvingSlotDoesNotWidenAnyOtherLookup(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	other := uuid.New()
	pushResolvingSpell(t, g, me.ID, func(g *Game, _ *StackItem) error {
		if _, _, ok := g.stackSpellLocked(other); ok {
			t.Error("a spell that is not on the stack and is not resolving was found")
		}
		if err := g.CopySpellForEffect(other, me.ID, false, nil); err != ErrCardNotFound {
			t.Errorf("CopySpellForEffect on a missing spell = %v, want ErrCardNotFound", err)
		}
		return nil
	})
	g.WithWriteLock(func() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			t.Fatalf("resolveTopOfStackLocked: %v", err)
		}
	})
}

// TestResolvingSlotRewindsWithAnUndo — the clone carries it, so an
// undo across the paused prompt can answer it again.
func TestResolvingSlotRewindsWithAnUndo(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	var id uuid.UUID
	found := 0
	id = pushResolvingSpell(t, g, me.ID, func(g *Game, item *StackItem) error {
		g.QueueConfirmForEffect(ConfirmPrompt{
			Chooser:  me.ID,
			Source:   item.SourceCardID,
			Question: "test — pause here",
			OnAccept: func(g *Game) error {
				if _, _, ok := g.stackSpellLocked(id); ok {
					found++
				}
				return nil
			},
		})
		return nil
	})
	g.WithWriteLock(func() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			t.Fatalf("resolveTopOfStackLocked: %v", err)
		}
	})
	snap := g.Clone()
	c := latestChoiceOfKindForTest(g, PendingChoiceConfirm)
	if err := g.ResolveConfirm(c.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if found != 1 {
		t.Fatalf("the source was reachable %d times, want 1", found)
	}
	g.RestoreFrom(snap)
	back := latestChoiceOfKindForTest(g, PendingChoiceConfirm)
	if back == nil {
		t.Fatalf("undo puts the prompt back: %+v", g.PendingChoices)
	}
	if err := g.ResolveConfirm(back.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm after undo: %v", err)
	}
	if found != 2 {
		t.Errorf("the source was reachable %d times, want 2 — the undo dropped the resolving item", found)
	}
}

// latestChoiceOfKindForTest returns the newest open choice of a kind.
func latestChoiceOfKindForTest(g *Game, kind PendingChoiceKind) *PendingChoice {
	var out *PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind {
			out = c
		}
	}
	return out
}

// TestResolvingSlotIsDroppedBySnapshotAndRestoresEmpty — the drift
// plan says dropped; this is what dropped looks like.
func TestResolvingSlotIsDroppedBySnapshotAndRestoresEmpty(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	pushResolvingSpell(t, g, me.ID, func(*Game, *StackItem) error { return nil })
	g.WithWriteLock(func() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			t.Fatalf("resolveTopOfStackLocked: %v", err)
		}
	})
	snap := g.CaptureSnapshot()
	back, err := snap.Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if back.resolving != nil {
		t.Errorf("the restored game carries a resolving item: %+v", back.resolving)
	}
}

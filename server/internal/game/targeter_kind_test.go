package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// targeter_kind_test.go — #3779: what kind of object did the targeting
// (Event.TargeterKind), and the per-player record of activations that
// targeted one of the activator's own creatures
// (ActivationTally.OwnCreatureTargeted).

func becameTargetEvents(g *Game) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventBecomesTarget {
			out = append(out, ev)
		}
	}
	return out
}

// TestBecomesTargetCarriesTheTargetersKind: one event per target slot,
// each stamped with the kind of the stack item that named it.
func TestBecomesTargetCarriesTheTargetersKind(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	victim := crimeTestCard(g, g.Battlefield, "Victim", me.ID, me.ID)
	other := crimeTestCard(g, g.Battlefield, "Other", me.ID, me.ID)

	g.WithWriteLock(func() {
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*StackItem{}
		}
		for _, k := range []StackItemKind{StackItemSpell, StackItemActivated, StackItemTriggered} {
			id := uuid.New()
			g.StackMeta[id] = &StackItem{ID: id, Kind: k, Controller: me.ID}
			g.emitBecameTargetLocked(me.ID, uuid.New(), id, []TargetRef{{Kind: TargetCard, ID: victim}, {Kind: TargetCard, ID: other}})
		}
		// An item that is not on the stack reads empty rather than guessing.
		g.emitBecameTargetLocked(me.ID, uuid.New(), uuid.New(), []TargetRef{{Kind: TargetCard, ID: victim}})
	})

	evs := becameTargetEvents(g)
	want := []StackItemKind{StackItemSpell, StackItemSpell, StackItemActivated, StackItemActivated, StackItemTriggered, StackItemTriggered, ""}
	if len(evs) != len(want) {
		t.Fatalf("%d became-target events, want %d", len(evs), len(want))
	}
	for i, ev := range evs {
		if ev.TargeterKind != want[i] {
			t.Errorf("event %d TargeterKind = %q, want %q", i, ev.TargeterKind, want[i])
		}
	}
}

// TestTargeterKindSurvivesCloneAndDiskRestore: the field rides the event
// log, so undo and a restart keep it.
func TestTargeterKindSurvivesCloneAndDiskRestore(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	victim := crimeTestCard(g, g.Battlefield, "Victim", me.ID, me.ID)
	g.WithWriteLock(func() {
		id := uuid.New()
		g.StackMeta = map[uuid.UUID]*StackItem{id: {ID: id, Kind: StackItemActivated, Controller: me.ID}}
		g.emitBecameTargetLocked(me.ID, uuid.New(), id, []TargetRef{{Kind: TargetCard, ID: victim}})
	})
	if got := becameTargetEvents(g.Clone()); len(got) != 1 || got[0].TargeterKind != StackItemActivated {
		t.Errorf("clone lost the kind: %+v", got)
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatal(err)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	if got := becameTargetEvents(restored); len(got) != 1 || got[0].TargeterKind != StackItemActivated {
		t.Errorf("disk restore lost the kind: %+v", got)
	}
}

// TestOwnCreatureTargetedRecord: counts an activation once however many
// of the activator's creatures it names, and not for an opponent's
// creature, a non-creature, or a player.
func TestOwnCreatureTargetedRecord(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := crimeTestCard(g, g.Battlefield, "Mine", me.ID, me.ID)
	mine2 := crimeTestCard(g, g.Battlefield, "Mine Too", me.ID, me.ID)
	theirs := crimeTestCard(g, g.Battlefield, "Theirs", opp.ID, opp.ID)
	rock := uuid.New()
	g.Battlefield.PushTop(Card{InstanceID: rock, Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})

	card := func(id uuid.UUID) TargetRef { return TargetRef{Kind: TargetCard, ID: id} }
	count := func() (n int) {
		g.ReadSnapshot(func() { n = g.OwnCreatureTargetedActivationsThisTurn(me.ID) })
		return
	}
	note := func(ts ...TargetRef) {
		g.WithWriteLock(func() { g.noteOwnCreatureTargetedLocked(me.ID, ts) })
	}

	note(card(theirs), card(rock), TargetRef{Kind: TargetPlayer, ID: opp.ID})
	if got := count(); got != 0 {
		t.Fatalf("an opponent's creature, an artifact and a player counted: %d", got)
	}
	note(card(mine), card(mine2))
	if got := count(); got != 1 {
		t.Errorf("one activation naming two of my creatures counted %d, want 1", got)
	}
	note(card(mine))
	if got := count(); got != 2 {
		t.Errorf("a second activation counted %d total, want 2", got)
	}
	g.ReadSnapshot(func() {
		if g.OwnCreatureTargetedActivationsThisTurn(opp.ID) != 0 {
			t.Error("the record leaked to another player")
		}
	})

	// A clone owns its map: a later write does not reach it.
	clone := g.Clone()
	note(card(mine))
	clone.ReadSnapshot(func() {
		if got := clone.OwnCreatureTargetedActivationsThisTurn(me.ID); got != 2 {
			t.Errorf("the clone saw a later write: %d, want 2", got)
		}
	})

	// The per-turn half empties on the turn advance.
	g.WithWriteLock(func() { g.resetActivationTurnTallyLocked() })
	if got := count(); got != 0 {
		t.Errorf("the record survived the turn advance: %d", got)
	}
}

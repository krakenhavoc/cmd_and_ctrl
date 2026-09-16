package game

import (
	"fmt"
	"testing"
)

// clone_events_test.go — #629: the undo snapshot shares the event log
// instead of deep-copying it, and RestoreFrom truncates to the
// length the snapshot recorded.
//
// The log is append-only and a snapshot never reads past its own
// length, which is what makes sharing sound. These tests pin the
// three things that has to mean: an undo really does drop the events
// the undone action emitted, the Seq stream rewinds with them, and
// one snapshot's restore can never rewrite entries another snapshot
// is still pointing at.

// emitFillerEvents appends n cheap events to the log.
func emitFillerEvents(g *Game, n int) {
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			g.EmitEvent(Event{Kind: EventDrawCard, Actor: g.Seats[0].ID})
		}
	})
}

// The headline case from the issue: clone at 10,000 events, append
// 100, restore, and the log is back to 10,000 with the Seq stream
// continuing from there.
func TestRestoreTruncatesTheEventLogAndRewindsSeq(t *testing.T) {
	g := newActiveGame(t)
	emitFillerEvents(g, 10_000)
	baseline := len(g.Events)
	baselineSeq := g.Events[baseline-1].Seq

	snapshot := g.Clone()
	emitFillerEvents(g, 100)
	if got := len(g.Events); got != baseline+100 {
		t.Fatalf("log is %d after the appends, want %d", got, baseline+100)
	}
	if got := len(snapshot.Events); got != baseline {
		t.Errorf("the snapshot grew to %d: it records its own length and must not see later appends", got)
	}

	g.WithWriteLock(func() { g.RestoreFrom(snapshot) })

	if got := len(g.Events); got != baseline {
		t.Errorf("log is %d after the restore, want %d", got, baseline)
	}
	if got := g.Events[len(g.Events)-1].Seq; got != baselineSeq {
		t.Errorf("last event Seq %d, want %d", got, baselineSeq)
	}
	emitFillerEvents(g, 1)
	if got := g.Events[len(g.Events)-1].Seq; got != baselineSeq+1 {
		t.Errorf("the next event has Seq %d, want %d — eventSeq rewinds with the log", got, baselineSeq+1)
	}
	if got := len(g.Events); got != baseline+1 {
		t.Errorf("log is %d after one more event, want %d", got, baseline+1)
	}
}

// The property the change is FOR: the snapshot's log is the live log,
// not a copy of it.
func TestCloneSharesTheEventLogsBackingArray(t *testing.T) {
	g := newActiveGame(t)
	emitFillerEvents(g, 64)
	snapshot := g.Clone()
	if &snapshot.Events[0] != &g.Events[0] {
		t.Error("the snapshot copied the event log; it is append-only and is meant to be shared (#629)")
	}
	if cap(snapshot.Events) != len(snapshot.Events) {
		t.Errorf("snapshot log cap %d, len %d: the cap must be pinned to the recorded length so nothing can write past it",
			cap(snapshot.Events), len(snapshot.Events))
	}
}

// The safety belt, stated as behaviour: restore an early snapshot,
// keep playing, and a later snapshot's view of the log is untouched.
// Without the capacity cap the new events would be written straight
// over the entries it is still pointing at.
func TestEventsAfterARestoreDoNotRewriteAnotherSnapshot(t *testing.T) {
	g := newActiveGame(t)
	emitFillerEvents(g, 8)
	early := g.Clone()
	emitFillerEvents(g, 8)
	later := g.Clone()

	laterSeqs := make([]uint64, len(later.Events))
	for i, ev := range later.Events {
		laterSeqs[i] = ev.Seq
	}

	g.WithWriteLock(func() { g.RestoreFrom(early) })
	emitFillerEvents(g, 8)

	for i, ev := range later.Events {
		if ev.Seq != laterSeqs[i] {
			t.Fatalf("event %d in the later snapshot changed Seq %d → %d after a restore and eight new events",
				i, laterSeqs[i], ev.Seq)
		}
	}
}

// TurnTally.FirstEvent (#591) indexes the log, so a truncation could
// leave it dangling. It cannot: the tally is restored from the same
// snapshot, so it points at most at the restored end.
func TestTurnTallyFirstEventSurvivesARestore(t *testing.T) {
	g := newActiveGame(t)
	emitFillerEvents(g, 32)
	snapshot := g.Clone()
	emitFillerEvents(g, 32)

	g.WithWriteLock(func() { g.RestoreFrom(snapshot) })

	if g.TurnTally.FirstEvent > len(g.Events) {
		t.Fatalf("FirstEvent %d points past the restored log (%d events)", g.TurnTally.FirstEvent, len(g.Events))
	}
	if n := len(g.EventsThisTurn()); n != len(g.Events)-g.TurnTally.FirstEvent {
		t.Errorf("EventsThisTurn returned %d, want %d", n, len(g.Events)-g.TurnTally.FirstEvent)
	}
}

// BenchmarkCloneWithEvents is the number #629 was opened about: the
// undo stack takes one Clone per action, so the cost of the log copy
// was paid on every single thing a player did.
func BenchmarkCloneWithEvents(b *testing.B) {
	for _, n := range []int{1_000, 10_000, 100_000} {
		b.Run(fmt.Sprintf("events=%d", n), func(b *testing.B) {
			g := newActiveGameWithSeats(b, 2)
			emitFillerEvents(g, n)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = g.Clone()
			}
		})
	}
}

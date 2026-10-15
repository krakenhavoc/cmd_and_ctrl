package game

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// freshEncoding is what json.Marshal writes for events as a []Event,
// without the brackets: the encoding the cache must reproduce.
func freshEncoding(t *testing.T, events []Event) []byte {
	t.Helper()
	if len(events) == 0 {
		return nil
	}
	b, err := json.Marshal(events)
	if err != nil {
		t.Fatal(err)
	}
	return b[1 : len(b)-1]
}

// emitVariedEvents emits n events with the fields that encode in
// interesting ways: strings that json escapes (HTML characters, quotes,
// non-ASCII), slices, ids and the omitempty zero values.
func emitVariedEvents(g *Game, n int) {
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			switch i % 4 {
			case 0:
				g.EmitEvent(Event{Kind: EventDrawCard, Actor: g.Seats[0].ID})
			case 1:
				g.EmitEvent(Event{Kind: EventStepBegan, Actor: g.Seats[1].ID, Amount: i, Label: `<b>"Step" & — "end"</b>`})
			case 2:
				g.EmitEvent(Event{Kind: EventChangeLife, Target: g.Seats[0].ID, Amount: -i, Colors: []string{"R", "G"}, Seats: []int{0, 1}})
			default:
				g.EmitEvent(Event{Kind: EventSearchLibrary, Actor: g.Seats[0].ID, CardID: uuid.New(), Label: "shuffle"})
			}
		}
	})
}

// TestEncodedEventLogMatchesMarshal is #3823's cache contract: what the
// cache returns is byte for byte what json.Marshal writes for the same
// events, as the log grows, for an older (shorter) capture, and after
// the log is replaced under a new generation (an undo) and regrows with
// different events at the same indices.
func TestEncodedEventLogMatchesMarshal(t *testing.T) {
	g := newActiveGame(t)
	var c EncodedEventLog

	check := func(step string) {
		t.Helper()
		g.mu.RLock()
		events, gen := g.Events[:len(g.Events):len(g.Events)], g.eventLogGen
		g.mu.RUnlock()
		got, err := c.Encode(events, gen)
		if err != nil {
			t.Fatalf("%s: Encode: %v", step, err)
		}
		if want := freshEncoding(t, events); !bytes.Equal(got, want) {
			t.Fatalf("%s: cached encoding differs from json.Marshal\n got: %.300s\nwant: %.300s", step, got, want)
		}
		if c.Len() != len(events) {
			t.Fatalf("%s: cache holds %d events, want %d", step, c.Len(), len(events))
		}
	}

	check("empty log")
	emitVariedEvents(g, 40)
	check("first encode")
	emitVariedEvents(g, 13)
	check("appended")
	check("unchanged")

	// An older capture of the same history is a prefix of the cache.
	g.mu.RLock()
	older := g.Events[:20:20]
	gen := g.eventLogGen
	g.mu.RUnlock()
	got, err := c.Encode(older, gen)
	if err != nil {
		t.Fatal(err)
	}
	if want := freshEncoding(t, older); !bytes.Equal(got, want) {
		t.Fatal("a shorter capture of the same history did not get its prefix")
	}

	// Undo: the log is truncated and regrows under a new generation with
	// different events at the indices the cache already encoded.
	before := g.Clone()
	emitVariedEvents(g, 10)
	check("before the undo")
	g.RestoreFrom(before)
	g.WithWriteLock(func() {
		for range 10 {
			g.EmitEvent(Event{Kind: EventChangeLife, Target: g.Seats[1].ID, Amount: 7, Label: "regrown"})
		}
	})
	check("regrown after the undo")
}

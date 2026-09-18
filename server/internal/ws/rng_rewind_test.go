package ws

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rng_rewind_test.go is ADR 0054 test plan item 13 for sub-PR 1: undo
// and a failed bundle rewind the game's random streams, so the same
// action redone draws the same result. Room needs no RNG code for
// this; the pre-action clone carries the key and the counters.

// libraryOrder reads a seat's library order under the game's read
// lock. PlayerByIDForEffect for the reason lifeIn gives (#877): a
// locking accessor inside a snapshot body is a recursive RLock.
func libraryOrder(g *game.Game, seat uuid.UUID) []string {
	var out []string
	g.ReadSnapshot(func() {
		if p := g.PlayerByIDForEffect(seat); p != nil {
			for _, c := range p.Library.Cards {
				out = append(out, c.Name)
			}
		}
	})
	return out
}

func shuffleVia(t *testing.T, room *Room, seat uuid.UUID) []string {
	t.Helper()
	if _, _, err := room.Apply(seat, func() error { return room.Game.ShuffleLibrary(seat) }); err != nil {
		t.Fatalf("Apply shuffle: %v", err)
	}
	return libraryOrder(room.Game, seat)
}

// TestUndoRewindsAShuffle: shuffle, undo, shuffle again gives the same
// library order. Before ADR 0054 the redo reshuffled.
func TestUndoRewindsAShuffle(t *testing.T) {
	room, _ := newBundleRoom(t)
	seat := room.Game.Seats[0].ID
	before := libraryOrder(room.Game, seat)

	first := shuffleVia(t, room, seat)
	if reflect.DeepEqual(first, before) {
		t.Fatal("the shuffle did not change the library order")
	}
	if _, _, err := room.Undo(uuid.Nil); err != nil {
		t.Fatalf("Undo: %v", err)
	}
	if got := libraryOrder(room.Game, seat); !reflect.DeepEqual(got, before) {
		t.Fatal("undo did not restore the pre-shuffle library")
	}
	if redo := shuffleVia(t, room, seat); !reflect.DeepEqual(redo, first) {
		t.Errorf("undo then redo reshuffled:\nfirst: %v\nredo:  %v", first, redo)
	}

	// Negative: WITHOUT an undo, the next shuffle is a new draw.
	if next := shuffleVia(t, room, seat); reflect.DeepEqual(next, first) {
		t.Error("a second shuffle with no undo repeated the first order")
	}
}

// TestFailedBundleRewindsTheRandomStream: a bundle whose shuffle step
// ran before a later step failed leaves the stream where it was, so
// the next shuffle matches a room in which the bundle never ran.
func TestFailedBundleRewindsTheRandomStream(t *testing.T) {
	clean, _ := newBundleRoom(t)
	want := shuffleVia(t, clean, clean.Game.Seats[0].ID)

	room, _ := newBundleRoom(t)
	g := room.Game
	seat := g.Seats[0].ID
	boom := errors.New("boom")
	_, _, err := room.ApplyBundle(Bundle{
		Steps: []func() error{
			func() error { return g.ShuffleLibrary(seat) },
			func() error { return boom },
		},
	})
	if !errors.Is(err, boom) {
		t.Fatalf("ApplyBundle err = %v, want the step's error", err)
	}
	if got := shuffleVia(t, room, seat); !reflect.DeepEqual(got, want) {
		t.Errorf("a rolled-back bundle's shuffle still advanced the stream:\ngot:  %v\nwant: %v", got, want)
	}
}

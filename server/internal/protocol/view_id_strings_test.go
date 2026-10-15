package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestIDStringsAnswerUUIDString pins the frame's seat-ID memo (#3823)
// to uuid.UUID.String: a seat, a non-seat, the nil UUID and a nil memo
// all read exactly as String() writes them, and a seat's string is the
// memo's own, not a fresh one per card.
func TestIDStringsAnswerUUIDString(t *testing.T) {
	seats := []*game.Player{{ID: uuid.New()}, {ID: uuid.New()}, nil, {ID: uuid.New()}}
	m := newIDStrings(seats)
	stranger := uuid.New()
	for _, id := range []uuid.UUID{seats[0].ID, seats[1].ID, seats[3].ID, stranger, uuid.Nil} {
		if got, want := m.of(id), id.String(); got != want {
			t.Errorf("memo.of(%s) = %q, want %q", id, got, want)
		}
		var none *idStrings
		if got, want := none.of(id), id.String(); got != want {
			t.Errorf("nil memo.of(%s) = %q, want %q", id, got, want)
		}
	}
	if allocs := testing.AllocsPerRun(100, func() { _ = m.of(seats[1].ID) }); allocs != 0 {
		t.Errorf("a seat's ID string allocated %.0f times per read; want 0 (the memo's own string)", allocs)
	}
}

// TestIDStringsHoldsATableOfSeats: every seat of a full table is in the
// memo, and a ninth is formatted rather than remembered.
func TestIDStringsHoldsATableOfSeats(t *testing.T) {
	var seats []*game.Player
	for range idStringsMax + 1 {
		seats = append(seats, &game.Player{ID: uuid.New()})
	}
	m := newIDStrings(seats)
	if len(m.ids) != idStringsMax {
		t.Fatalf("memo holds %d seats, want the cap %d", len(m.ids), idStringsMax)
	}
	last := seats[len(seats)-1].ID
	if got := m.of(last); got != last.String() {
		t.Errorf("an unremembered seat read %q, want %q", got, last.String())
	}
}

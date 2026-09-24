package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// put_in_library_test.go — ADR 0088's ordered placement is a prompt a
// seat owes the table, so the enumerator must offer it answers, every
// answer it offers must be one the dispatcher and resolver accept, and
// every other seat must be offered nothing while it is open.

func TestPutInLibraryEveryPlacementIsAnswerable(t *testing.T) {
	for _, placement := range []game.LibraryPlacement{
		game.LibraryPlaceTop, game.LibraryPlaceBottom, game.LibraryPlaceTopOrBottom,
	} {
		t.Run(string(placement), func(t *testing.T) {
			g := newTable(t)
			me, them := g.Seats[0], g.Seats[1]
			ids := make([]uuid.UUID, 0, 3)
			for _, name := range []string{"A", "B", "C"} {
				id := uuid.New()
				ids = append(ids, id)
				me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
			}
			g.WithWriteLock(func() {
				if err := g.PutInLibraryInChosenOrderThenForEffect(game.PutInLibrarySpec{
					Chooser: me.ID, Cards: ids, From: game.ZoneHand, Placement: placement,
				}); err != nil {
					t.Fatalf("queue: %v", err)
				}
			})

			moves := legal.EnumerateFor(g, me.ID)
			if len(moves) == 0 {
				t.Fatal("the seat owing the order was offered nothing (#499 / #618)")
			}
			for _, m := range moves {
				if m.Type != legal.TypeResolveChoice {
					t.Errorf("offered %q alongside the answers", m.Label)
				}
			}
			// Leave it alone, and each of the other two to the front.
			want := 3
			if placement == game.LibraryPlaceTopOrBottom {
				// … plus all on the bottom, and each one on the bottom.
				want += 1 + 3
			}
			if len(moves) != want {
				t.Errorf("%d answers, want %d: %v", len(moves), want, labels(moves))
			}
			if others := legal.EnumerateFor(g, them.ID); len(others) != 0 {
				t.Errorf("another seat was offered %v while the order is open", labels(others))
			}
			dispatchAll(t, g, me.ID, moves)
		})
	}
}

package protocol

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// final_board_view_test.go — #4912: an eliminated seat's final board on
// the wire. Absent while the player plays; once they have left, every
// viewer gets the same board, hand included (owner decision 2026-11-04),
// and never the library.

func newThreeSeatGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 3; i++ {
		deck := make([]game.Card, 10)
		for j := range deck {
			deck[j] = game.NewCard("filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer %d: %v", i, err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(5, 6))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

func finalCardNamed(cards []FinalCardView, name string) (FinalCardView, bool) {
	for _, c := range cards {
		if c.Name == name {
			return c, true
		}
	}
	return FinalCardView{}, false
}

func TestEliminatedSeatCarriesItsFinalBoardForEveryViewer(t *testing.T) {
	g := newThreeSeatGame(t)
	a, b, c := g.Seats[0], g.Seats[1], g.Seats[2]
	const libraryCard = "B's Library Card"
	g.WithWriteLock(func() {
		perm := game.NewCard("B's Bear", b.ID)
		perm.TypeLine = "Creature — Bear"
		perm.ScryfallID = "scry-bear"
		perm.Tapped = true
		perm.Counters = map[string]int{game.CounterPlusOne: 1}
		g.Battlefield.PushTop(perm)
		b.Hand.PushTop(game.NewCard("B's Last Card", b.ID))
		b.Graveyard.PushTop(game.NewCard("B's Grave Card", b.ID))
		b.Library.Cards[0].Name = libraryCard
		b.Command.PushTop(game.NewCommander("B's Commander", b.ID))
	})

	// While B plays, nobody's seat has a final board.
	for _, s := range ViewOfGame(g).Seats {
		if s.FinalBoard != nil {
			t.Fatalf("seat %s has a final board while everyone is in", s.Name)
		}
	}

	if err := g.Concede(b.ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	v := ViewOfGame(g)
	for _, viewer := range []string{a.ID.String(), b.ID.String(), c.ID.String(), ""} {
		fv := FilterViewFor(v, viewer)
		fb := fv.Seats[b.Seat].FinalBoard
		if fb == nil {
			t.Fatalf("viewer %q: the eliminated seat has no final board", viewer)
		}
		bear, ok := finalCardNamed(fb.Battlefield, "B's Bear")
		if !ok || !bear.Tapped || bear.Counters[game.CounterPlusOne] != 1 || bear.ScryfallID != "scry-bear" {
			t.Errorf("viewer %q: final battlefield %+v", viewer, fb.Battlefield)
		}
		if _, ok := finalCardNamed(fb.Hand, "B's Last Card"); !ok {
			t.Errorf("viewer %q: the last hand is not shown once B has left: %+v", viewer, fb.Hand)
		}
		if _, ok := finalCardNamed(fb.Graveyard, "B's Grave Card"); !ok {
			t.Errorf("viewer %q: final graveyard %+v", viewer, fb.Graveyard)
		}
		// #4547: the tombstone's commander line, the same for every viewer.
		if len(fb.Commanders) != 1 || fb.Commanders[0] != "B's Commander" {
			t.Errorf("viewer %q: final board commanders %q, want B's Commander", viewer, fb.Commanders)
		}
		for _, zone := range [][]FinalCardView{fb.Battlefield, fb.Hand, fb.Graveyard, fb.Exile, fb.Command} {
			if _, ok := finalCardNamed(zone, libraryCard); ok {
				t.Errorf("viewer %q: a library card is on the final board", viewer)
			}
		}
		for _, s := range fv.Seats {
			if s.ID != b.ID.String() && s.FinalBoard != nil {
				t.Errorf("viewer %q: seat %s, still in, has a final board", viewer, s.Name)
			}
		}
	}
}

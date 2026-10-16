package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// outside_play_view_test.go — #3955, Wish: the seat's right to play a card
// from outside the game is its owner's to see, and says whether the
// `outside_play` action is how it names the card.

type emptyViewCollection struct{}

func (emptyViewCollection) Search(string, int) []game.Card { return nil }
func (emptyViewCollection) Named(string) (game.Card, bool) { return game.Card{}, false }
func (emptyViewCollection) Suggest(func(game.Card) bool)   {}

func seatOutsidePlay(v GameView, seat string) *OutsidePlayView {
	for _, s := range v.Seats {
		if s.ID == seat {
			return s.OutsidePlay
		}
	}
	return nil
}

func TestTheWishRightIsTheOwnersAndNobodyElses(t *testing.T) {
	prev := game.OutsideCollection
	game.OutsideCollection = emptyViewCollection{}
	t.Cleanup(func() { game.OutsideCollection = prev })

	g, _ := outsideCastGame(t, "Land", "") // seat 0 has a sideboard, seat 1 none
	a, b := g.Seats[0].ID, g.Seats[1].ID
	if got := seatOutsidePlay(ViewOfGameFor(g, a.String()), a.String()); got != nil {
		t.Fatalf("a seat with no Wish sees a right: %+v", got)
	}
	g.WithWriteLock(func() {
		g.GrantOutsidePlayRightForEffect(a, uuid.New(), "Wish")
		g.GrantOutsidePlayRightForEffect(b, uuid.New(), "Wish")
	})

	// A sideboard pool plays its own cards directly: no naming.
	if got := seatOutsidePlay(ViewOfGameFor(g, a.String()), a.String()); got == nil || got.SourceName != "Wish" || got.CanName {
		t.Errorf("sideboard pool: %+v, want the right without can_name", got)
	}
	// The database pool names its card.
	if got := seatOutsidePlay(ViewOfGameFor(g, b.String()), b.String()); got == nil || !got.CanName {
		t.Errorf("database pool: %+v, want can_name", got)
	}
	// Nobody else sees it.
	for _, viewer := range []string{b.String(), SpectatorViewerID, ""} {
		if got := seatOutsidePlay(ViewOfGameFor(g, viewer), a.String()); got != nil {
			t.Errorf("viewer %q sees seat 0's right: %+v", viewer, got)
		}
	}
}

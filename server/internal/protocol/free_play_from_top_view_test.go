package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// free_play_from_top_view_test.go — #3801. "Look at the top card of your
// library, you may play it free" names a card only its looker has seen
// (CR 701.20). The may_cast offer's words and its card id are the
// chooser's alone; the table sees that a seat owes an answer and nothing
// about which card. A reveal (Rashmi) is public.

func topPlayOfferTable(t *testing.T, reveal bool) (g *game.Game, me, them *game.Player, top game.Card) {
	t.Helper()
	g = buildActiveGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	them = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	withTopVisibility(t, "unused", game.LibraryTopHidden)
	top = game.NewCard("Secret Spell", me.ID)
	top.TypeLine = "Instant"
	top.ManaCost = "{1}"
	me.Library.PushTop(top)
	g.WithWriteLock(func() {
		if err := g.OfferFreePlayFromTopOrHandForEffect(game.FreePlayFromTop{
			Player: me.ID, Source: uuid.New(), SourceName: "Test Envoy",
			MaxManaValue: 3, Land: true, Reveal: reveal,
		}); err != nil {
			t.Fatalf("offer: %v", err)
		}
	})
	return g, me, them, top
}

func mayCastFor(g *game.Game, viewer *game.Player) *PendingChoiceView {
	v := ViewOfGameFor(g, viewer.ID.String())
	for i := range v.PendingChoices {
		if v.PendingChoices[i].Kind == string(game.PendingChoiceMayCast) {
			return &v.PendingChoices[i]
		}
	}
	return nil
}

func TestALookedAtTopCardOfferIsPrivateToItsChooser(t *testing.T) {
	g, me, them, top := topPlayOfferTable(t, false)

	own := mayCastFor(g, me)
	if own == nil {
		t.Fatal("the chooser has no may_cast offer")
	}
	if own.MayCastCard != top.InstanceID.String() || !strings.Contains(own.Reason, "Secret Spell") {
		t.Errorf("the chooser must see the card: card=%q reason=%q", own.MayCastCard, own.Reason)
	}

	other := mayCastFor(g, them)
	if other == nil {
		t.Fatal("the table does not see that a seat owes an answer")
	}
	if other.MayCastCard != "" || strings.Contains(other.Reason, "Secret Spell") ||
		strings.Contains(other.Reason, "Test Envoy") || other.AcceptLabel != "" || other.Source != "" {
		t.Errorf("an opponent's copy leaks the looked-at card: %+v", other)
	}
	if !other.PrivateText {
		t.Error("the opponent's copy is not marked private")
	}
	// The spectator and the admin view are held to the same rule.
	for _, viewer := range []string{SpectatorViewerID} {
		for _, c := range ViewOfGameFor(g, viewer).PendingChoices {
			if c.Kind == string(game.PendingChoiceMayCast) && (c.MayCastCard != "" || strings.Contains(c.Reason, "Secret Spell")) {
				t.Errorf("%q sees the looked-at card: %+v", viewer, c)
			}
		}
	}
	// And the card itself stays out of the opponent's view of the library.
	if card := cardInZone(ViewOfGameFor(g, them.ID.String()).Seats[seatIndex(g, me)].Library, top.InstanceID); card != nil && card.Name != "" {
		t.Errorf("the opponent can read the top card of the library: %+v", card)
	}
}

func TestARevealedTopCardOfferIsPublic(t *testing.T) {
	g, _, them, top := topPlayOfferTable(t, true)
	other := mayCastFor(g, them)
	if other == nil || other.MayCastCard != top.InstanceID.String() || other.PrivateText {
		t.Fatalf("a revealed card's offer must be public: %+v", other)
	}
}

// Accepting stamps the free cast on the looked-at card, so its controller's
// pile shows it castable, and no other seat gets the card or the button.
func TestAnAcceptedTopCardCastIsCastableForItsController(t *testing.T) {
	g, me, them, top := topPlayOfferTable(t, false)
	c := mayCastFor(g, me)
	if c == nil {
		t.Fatal("no offer")
	}
	id, err := uuid.Parse(c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.ResolveMayCast(id, me.ID, true); err != nil {
		t.Fatalf("ResolveMayCast: %v", err)
	}
	card := cardInZone(ViewOfGameFor(g, me.ID.String()).Seats[seatIndex(g, me)].Library, top.InstanceID)
	if card == nil || card.Name == "" || !card.CastableHere {
		t.Fatalf("the controller's library top is not a castable, readable card: %+v", card)
	}
	if theirs := cardInZone(ViewOfGameFor(g, them.ID.String()).Seats[seatIndex(g, me)].Library, top.InstanceID); theirs != nil && (theirs.Name != "" || theirs.CastableHere) {
		t.Errorf("the opponent's copy of the pile leaks the card: %+v", theirs)
	}
}

func seatIndex(g *game.Game, p *game.Player) int {
	for i, s := range g.Seats {
		if s == p {
			return i
		}
	}
	return 0
}

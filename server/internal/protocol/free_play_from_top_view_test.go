package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// free_play_from_top_view_test.go — #3801. "Look at the top card of your
// library, you may cast it free" names a card only its looker has seen
// (CR 701.20). Since #4250 the spell's offer is a resolution_cast made
// inside the resolution that looks: its card, still on top of the library,
// is the chooser's alone (the knower filter drops it from every other
// viewer's copy of the prompt), and the table sees that a seat owes an
// answer and what made the offer, and nothing about which card. The land's
// may_cast is private text. A reveal (Rashmi) is public.

// topPlayOfferTable makes the offer over a top card. The engine raises the
// resolution_cast only inside an open resolution; here the offer is made
// directly, so the permission is stamped and the prompt the resolution
// would raise is queued by hand, as resolution_cast_view_test.go does.
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
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind: game.PendingChoiceResolutionCast, Chooser: me.ID, Count: 1,
			Reason: "Test Envoy", DiscardOptions: []uuid.UUID{top.InstanceID},
			MayCastKeyword: game.MayCastKeywordFreeOrHand,
			AcceptLabel:    "Cast it free", DeclineLabel: "Put it into your hand",
		})
	})
	return g, me, them, top
}

func topOfferFor(g *game.Game, viewer string) *PendingChoiceView {
	v := ViewOfGameFor(g, viewer)
	for i := range v.PendingChoices {
		if v.PendingChoices[i].Kind == string(game.PendingChoiceResolutionCast) {
			return &v.PendingChoices[i]
		}
	}
	return nil
}

func TestALookedAtTopCardOfferIsPrivateToItsChooser(t *testing.T) {
	g, me, them, top := topPlayOfferTable(t, false)

	own := topOfferFor(g, me.ID.String())
	if own == nil {
		t.Fatal("the chooser has no offer")
	}
	if len(own.Options) != 1 || own.Options[0].InstanceID != top.InstanceID.String() || own.Options[0].Name != "Secret Spell" {
		t.Errorf("the chooser must see the card: %+v", own.Options)
	}
	if own.MayCastKeyword != game.MayCastKeywordFreeOrHand || own.DeclineLabel != "Put it into your hand" {
		t.Errorf("the chooser's words: keyword %q decline %q", own.MayCastKeyword, own.DeclineLabel)
	}

	other := topOfferFor(g, them.ID.String())
	if other == nil {
		t.Fatal("the table does not see that a seat owes an answer")
	}
	if len(other.Options) != 0 || len(other.Eligible) != 0 || strings.Contains(other.Reason, "Secret Spell") {
		t.Errorf("an opponent's copy leaks the looked-at card: %+v", other)
	}
	// The spectator is held to the same rule.
	if c := topOfferFor(g, SpectatorViewerID); c != nil && (len(c.Options) != 0 || len(c.Eligible) != 0) {
		t.Errorf("a spectator sees the looked-at card: %+v", c)
	}
	// And the card itself stays out of the opponent's view of the library.
	if card := cardInZone(ViewOfGameFor(g, them.ID.String()).Seats[seatIndex(g, me)].Library, top.InstanceID); card != nil && card.Name != "" {
		t.Errorf("the opponent can read the top card of the library: %+v", card)
	}
}

func TestARevealedTopCardOfferIsPublic(t *testing.T) {
	g, _, them, top := topPlayOfferTable(t, true)
	other := topOfferFor(g, them.ID.String())
	if other == nil || len(other.Options) != 1 || other.Options[0].InstanceID != top.InstanceID.String() {
		t.Fatalf("a revealed card's offer must be public: %+v", other)
	}
}

// The offer stamps the free cast on the looked-at card, so its
// controller's pile shows it readable and castable, and no other seat gets
// the card or the button.
func TestAnAcceptedTopCardCastIsCastableForItsController(t *testing.T) {
	g, me, them, top := topPlayOfferTable(t, false)
	if topOfferFor(g, me.ID.String()) == nil {
		t.Fatal("no offer")
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

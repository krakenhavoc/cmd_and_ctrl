package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hand_cast_offer_view_test.go — #4250, part 2. "You may cast a spell from
// your hand without paying its mana cost" is a resolution_cast over HAND
// cards. The prompt's options are the chooser's own hand, so the chooser
// sees each card and the table sees only that a seat owes an answer: the
// knower filter drops a hand card from every viewer who does not know it.
//
// The engine raises the prompt only inside an open resolution; here the
// permission is stamped and the prompt the resolution would raise is
// queued by hand, as free_play_from_top_view_test.go does.
func TestAHandCastOfferShowsTheHandToItsChooserAlone(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	c := game.NewCard("Secret Sorcery", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{3}{R}"
	c.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(c)
	g.WithWriteLock(func() {
		if !g.FreeCastFromHandForEffect(me.ID, uuid.Nil, c.InstanceID, "Test Sword") {
			t.Fatal("the grant found no card in hand")
		}
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind: game.PendingChoiceResolutionCast, Chooser: me.ID, Count: 1,
			Reason: "Test Sword", DiscardOptions: []uuid.UUID{c.InstanceID},
			MayCastKeyword: game.MayCastKeywordHandCast,
			AcceptLabel:    "Cast it free", DeclineLabel: "Keep it in your hand",
		})
	})

	own := topOfferFor(g, me.ID.String())
	if own == nil || len(own.Options) != 1 || own.Options[0].Name != "Secret Sorcery" ||
		len(own.Eligible) != 1 || own.Eligible[0] != c.InstanceID.String() {
		t.Fatalf("the chooser must see their hand card on the offer: %+v", own)
	}
	if own.MayCastKeyword != game.MayCastKeywordHandCast || own.DeclineLabel != "Keep it in your hand" {
		t.Errorf("the chooser's words: keyword %q decline %q", own.MayCastKeyword, own.DeclineLabel)
	}
	other := topOfferFor(g, them.ID.String())
	if other == nil {
		t.Fatal("the table does not see that a seat owes an answer")
	}
	if len(other.Options) != 0 || len(other.Eligible) != 0 {
		t.Errorf("an opponent's copy of the offer shows the hand card: %+v", other)
	}
}

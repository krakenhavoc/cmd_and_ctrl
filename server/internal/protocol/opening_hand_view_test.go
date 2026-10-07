package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Leyline of Sanctity
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// opening_hand_view_test.go — ADR 0133. The question an opening-hand
// action asks names a card in the chooser's hand, and the table has to
// see who the game is waiting on without seeing what is in that hand.

func openingOfferViews(t *testing.T) (chooser, other, spectator PendingChoiceView, leyline uuid.UUID) {
	t.Helper()
	g := buildActiveGame(t)
	me, opp := g.Seats[1], g.Seats[0]
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		c := game.NewCard("Leyline of Sanctity", me.ID)
		c.TypeLine = "Enchantment"
		c.ManaCost = "{2}{W}{W}"
		c.OracleID = "492e0e6c-8c27-4376-938b-f8a8b6205810"
		c.KnownBy = map[uuid.UUID]bool{me.ID: true}
		me.Hand.PushTop(c)
		leyline = c.InstanceID
	})
	for _, p := range []*game.Player{opp, me} {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	pick := func(viewer string) PendingChoiceView {
		t.Helper()
		for _, c := range ViewOfGameFor(g, viewer).PendingChoices {
			return c
		}
		t.Fatalf("viewer %q sees no pending choice", viewer)
		return PendingChoiceView{}
	}
	return pick(me.ID.String()), pick(opp.ID.String()), pick(""), leyline
}

func TestOpeningHandOfferNamesTheCardToItsChooserOnly(t *testing.T) {
	chooser, other, spectator, leyline := openingOfferViews(t)

	if !chooser.PrivateText || !strings.Contains(chooser.Reason, "Leyline of Sanctity") || chooser.Source != leyline.String() ||
		chooser.AcceptLabel == "" || chooser.DeclineLabel == "" {
		t.Errorf("the chooser lost the words of its own question: %+v", chooser)
	}
	for who, v := range map[string]PendingChoiceView{"another player": other, "a spectator": spectator} {
		if v.Kind != "confirm" || v.Chooser != chooser.Chooser {
			t.Errorf("%s no longer sees who the game is waiting on: kind %q chooser %q", who, v.Kind, v.Chooser)
		}
		if v.Source != "" || strings.Contains(v.Reason, "Leyline") || v.AcceptLabel != "" || v.DeclineLabel != "" {
			t.Errorf("%s was told what is in an opening hand: %+v", who, v)
		}
		if v.Reason == "" {
			t.Errorf("%s gets a prompt with no words at all", who)
		}
	}
}

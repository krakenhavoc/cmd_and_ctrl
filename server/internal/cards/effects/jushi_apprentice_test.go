package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func jushiApprenticeRow() cards.Card {
	return flipRow(jushiApprenticeOracleID, "{1}{U}", []string{"U"}, nil,
		cards.CardFace{Name: "Jushi Apprentice", TypeLine: "Creature — Human Wizard",
			OracleText: "{2}{U}, {T}: Draw a card. If you have nine or more cards in hand, flip this creature.",
			Power:      "1", Toughness: "2"},
		cards.CardFace{Name: "Tomoya the Revealer", TypeLine: "Legendary Creature — Human Wizard",
			OracleText: "{3}{U}{U}, {T}: Target player draws X cards, where X is the number of cards in your hand.",
			Power:      "2", Toughness: "3"})
}

// Jushi Apprentice draws, then flips if the hand (counted after the
// draw) holds nine or more.
func TestJushiApprenticeFlipsAtNineCards(t *testing.T) {
	for _, tc := range []struct {
		name     string
		hand     int // cards in hand before the draw
		wantFlip bool
	}{{"eight in hand", 8, true}, {"seven in hand", 7, false}, {"nine in hand", 9, true}} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := flipMain(t)
			id := flipEntered(t, g, jushiApprenticeRow(), me)
			g.WithWriteLock(func() { me.Hand.Cards = nil })
			for i := 0; i < tc.hand; i++ {
				handCardForTest(me, "Filler", "Sorcery", "")
			}
			p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if me.Hand.Size() != tc.hand+1 {
				t.Fatalf("hand is %d after the draw, want %d", me.Hand.Size(), tc.hand+1)
			}
			if tc.wantFlip {
				wantFlipped(t, g, id, "Tomoya the Revealer", []string{"U"}, 2)
			} else {
				wantUnflipped(t, g, id, "Jushi Apprentice")
			}
		})
	}
}

// Tomoya's X is the number of cards in the CONTROLLER's hand, whoever
// the target is.
func TestTomoyaDrawsAsManyAsYourHand(t *testing.T) {
	g, me, opp := flipMain(t)
	id := flipEntered(t, g, jushiApprenticeRow(), me)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
		me.Hand.Cards = nil
		opp.Hand.Cards = nil
	})
	for i := 0; i < 3; i++ {
		handCardForTest(me, "Filler", "Sorcery", "")
	}
	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Hand.Size() != 3 {
		t.Errorf("the opponent drew to %d cards, want 3 (my hand size)", opp.Hand.Size())
	}
	if me.Hand.Size() != 3 {
		t.Errorf("my hand is %d, want unchanged 3", me.Hand.Size())
	}
}

package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func budokaGardenerRow() cards.Card {
	return flipRow(budokaGardenerOracleID, "{1}{G}", []string{"G"}, nil,
		cards.CardFace{Name: "Budoka Gardener", TypeLine: "Creature — Human Monk",
			OracleText: "{T}: You may put a land card from your hand onto the battlefield. If you control ten or more lands, flip this creature.",
			Power:      "2", Toughness: "1"},
		cards.CardFace{Name: "Dokai, Weaver of Life", TypeLine: "Legendary Creature — Human Monk",
			OracleText: "{4}{G}{G}, {T}: Create an X/X green Elemental creature token, where X is the number of lands you control.",
			Power:      "3", Toughness: "3"})
}

func pushForests(g *game.Game, p *game.Player, n int) {
	for i := 0; i < n; i++ {
		apaPush(g, p.ID, p.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	}
}

// Budoka Gardener flips when, after its land drop, its controller has
// ten or more lands: nine plus the land from hand flips it, eight plus
// one does not, and the count is asked even when the land was declined
// or there was none to put.
func TestBudokaGardenerFlipsAtTenLands(t *testing.T) {
	for _, tc := range []struct {
		name     string
		lands    int  // already on the battlefield
		inHand   bool // a land card to put
		put      bool // whether the player puts it
		wantFlip bool
	}{
		{"nine plus the land from hand", 9, true, true, true},
		{"eight plus the land from hand", 8, true, true, false},
		{"nine, land declined", 9, true, false, false},
		{"ten, land declined", 10, true, false, true},
		{"ten, no land in hand", 10, false, false, true},
		{"nine, no land in hand", 9, false, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := flipMain(t)
			id := flipEntered(t, g, budokaGardenerRow(), me)
			pushForests(g, me, tc.lands)
			var land uuid.UUID
			if tc.inHand {
				land = handCardForTest(me, "Forest", "Basic Land — Forest", "")
			}
			p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
			if tc.inHand {
				if tc.put {
					answerChooseCards(t, g, me.ID, land)
				} else {
					answerChooseCards(t, g, me.ID)
				}
				passPriorityAroundTable(t, g)
			}
			if tc.wantFlip {
				wantFlipped(t, g, id, "Dokai, Weaver of Life", []string{"G"}, 2)
			} else {
				wantUnflipped(t, g, id, "Budoka Gardener")
			}
			if tc.put && !g.Battlefield.Contains(land) {
				t.Error("the land did not reach the battlefield")
			}
		})
	}
}

// Dokai makes an X/X green Elemental, X the number of lands.
func TestDokaiMakesAnElementalAsBigAsYourLands(t *testing.T) {
	g, me, _ := flipMain(t)
	id := flipEntered(t, g, budokaGardenerRow(), me)
	pushForests(g, me, 7)
	g.WithWriteLock(func() {
		if err := g.FlipPermanentForEffect(id); err != nil {
			t.Fatal(err)
		}
	})
	wantFlipped(t, g, id, "Dokai, Weaver of Life", []string{"G"}, 2)

	p7Activate(t, g, me, id, 0, game.ActivateAbilityParams{})
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Elemental" {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil {
		t.Fatal("no Elemental token")
	}
	if e := token.Effective(); e.Power != 7 || e.Toughness != 7 || token.Controller != me.ID {
		t.Errorf("Elemental is %d/%d under %v, want 7/7 under me", e.Power, e.Toughness, token.Controller)
	}
	if c := token.Effective().Colors; len(c) != 1 || c[0] != "G" {
		t.Errorf("Elemental colours %v, want green", c)
	}
}

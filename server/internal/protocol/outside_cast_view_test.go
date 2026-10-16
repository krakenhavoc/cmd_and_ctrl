package protocol

import (
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// outside_cast_view_test.go — #3955: a card its owner may cast or play
// from outside the game carries the same cast surface a graveyard or
// exile card does, on the owner's frame and nobody else's.

func outsideCastGame(t *testing.T, typeLine, cost string) (*game.Game, game.Card) {
	t.Helper()
	g := game.NewGame()
	var card game.Card
	for i := 0; i < 2; i++ {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i), uuid.Nil)}
		for j := 0; j < 20; j++ {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if i == 0 {
			card = game.NewCard("Outside Test Card", uuid.Nil)
			card.TypeLine = typeLine
			card.ManaCost = cost
			card.IsSideboard = true
			deck = append(deck, card)
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Seat%d", i), deck); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.StartWithSource(nil); err != nil {
		t.Fatal(err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	return g, g.Seats[0].SideboardCards()[0]
}

func outsideView(t *testing.T, g *game.Game, viewer string) CardView {
	t.Helper()
	v := ViewOfGameFor(g, viewer)
	for _, s := range v.Seats {
		if s.ID == g.Seats[0].ID.String() && len(s.Sideboard) > 0 {
			return s.Sideboard[0]
		}
	}
	return CardView{}
}

func TestOutsideCardWithAFreeCastGrantIsCastableOnTheOwnersFrame(t *testing.T) {
	g, card := outsideCastGame(t, "Creature — Eldrazi", "{10}")
	owner, opp := g.Seats[0].ID, g.Seats[1].ID

	if got := outsideView(t, g, owner.String()); got.CastableHere {
		t.Fatalf("a sideboard card with no permission is castable_here: %+v", got)
	}

	g.WithWriteLock(func() { g.FreeCastFromOutsideForEffect(owner, uuid.New(), card.InstanceID, "Test Source") })
	got := outsideView(t, g, owner.String())
	if !got.CastableHere {
		t.Fatalf("the granted card is not castable_here: %+v", got)
	}
	if got.ExilePlay == nil || got.ExilePlay.CostOverride != "{0}" || !got.ExilePlay.CastOnly || !got.ExilePlay.ClosesOnPass {
		t.Errorf("exile_play = %+v, want a {0} cast-only pass-closed grant", got.ExilePlay)
	}
	if len(got.CastPrices) == 0 {
		t.Errorf("no cast price on a castable sideboard card")
	}
	if v := ViewOfGameFor(g, opp.String()); len(v.Seats[0].Sideboard) != 0 {
		t.Errorf("the opponent sees %d sideboard cards", len(v.Seats[0].Sideboard))
	}
	if raw := outsideView(t, g, SpectatorViewerID); raw.CastableHere || raw.InstanceID != "" {
		t.Errorf("the spectator sees the sideboard card: %+v", raw)
	}
}

func TestOutsideLandUnderAPlayGrantIsPlayable(t *testing.T) {
	g, _ := outsideCastGame(t, "Land", "")
	owner := g.Seats[0].ID
	g.WithWriteLock(func() { g.GrantOutsidePlayRightForEffect(owner, uuid.New(), "Test Wish") })
	if got := outsideView(t, g, owner.String()); !got.CastableHere {
		t.Errorf("a land under a play grant in the main phase is not castable_here: %+v", got)
	}
}

func TestOutsideLandUnderAFreeCastGrantIsNotPlayable(t *testing.T) {
	g, card := outsideCastGame(t, "Land", "")
	owner := g.Seats[0].ID
	g.WithWriteLock(func() { g.FreeCastFromOutsideForEffect(owner, uuid.New(), card.InstanceID, "Test Source") })
	if got := outsideView(t, g, owner.String()); got.CastableHere {
		t.Errorf("a land under a cast-only grant is castable_here: %+v", got)
	}
}

package legal_test

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// planar_die_test.go — ADR 0216 §4: the planar die is offered exactly
// when the engine accepts it (#544): the active player, a main phase, an
// empty stack, a cost they can pay, and no card.

type legalPlanes struct{}

func (legalPlanes) PlanarCards(game.PlanarKinds) []game.Card {
	out := make([]game.Card, 40)
	for i := range out {
		c := game.NewCard(fmt.Sprintf("Legal Plane %02d", i), uuid.Nil)
		c.OracleID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(c.Name)).String()
		c.TypeLine = "Plane — Legalworld"
		c.Layout = "planar"
		out[i] = c
	}
	return out
}

func planarDieMoves(moves []legal.Move) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Type != legal.TypeSpecialAction {
			continue
		}
		var p struct {
			Kind   string `json:"kind"`
			CardID string `json:"card_id"`
		}
		if json.Unmarshal(m.Params, &p) == nil && p.Kind == string(game.SpecialActionPlanarDie) {
			out = append(out, m)
		}
	}
	return out
}

func TestThePlanarDieIsOfferedWhenTheEngineTakesIt(t *testing.T) {
	prev := game.PlanarCards
	game.PlanarCards = legalPlanes{}
	t.Cleanup(func() { game.PlanarCards = prev })
	g := game.NewGame()
	for i := range 4 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Commander %d", i), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d-%d", i, j), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Seat%d", i), deck); err != nil {
			t.Fatal(err)
		}
	}
	on := true
	if err := g.UpdateSettings(uuid.Nil, game.SettingsPatch{Planechase: &on}); err != nil {
		t.Fatal(err)
	}
	if err := g.Start(rand.New(rand.NewPCG(7, 9))); err != nil {
		t.Fatal(err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	advanceTo(t, g, game.StepPrecombatMain)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]

	moves := planarDieMoves(legal.EnumerateFor(g, active.ID))
	if len(moves) != 1 {
		t.Fatalf("%d planar die moves at {0}, want 1", len(moves))
	}
	var p struct {
		CardID string `json:"card_id"`
	}
	_ = json.Unmarshal(moves[0].Params, &p)
	if p.CardID != "" || moves[0].Label != "Roll the planar die {0}" {
		t.Errorf("move %q names card %q", moves[0].Label, p.CardID)
	}
	if n := len(planarDieMoves(legal.EnumerateFor(g, other.ID))); n != 0 {
		t.Errorf("the non-active seat is offered the die %d times", n)
	}
	// The engine takes the move it was offered.
	if err := g.PerformSpecialAction(active.ID, uuid.Nil, game.SpecialActionPlanarDie, game.SpecialActionParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("the offered roll was refused: %v", err)
	}
	// {1} now, and the seat has no mana: not offered.
	g.WithWriteLock(func() { g.PendingTriggers = nil })
	if len(g.StackMeta) == 0 {
		if n := len(planarDieMoves(legal.EnumerateFor(g, active.ID))); n != 0 {
			t.Errorf("a {1} roll with no mana is offered %d times", n)
		}
	}
}

package aiseat_test

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // the opening-hand cards
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// opening_hand_test.go — ADR 0133. The questions an opening-hand action
// asks are the chained-choice kinds, which a bot seat answers with the
// rest of its choices; a seat owing one is offered nothing else, so a
// policy that could not answer it would hold the table (#544). This
// plays the opening to the end with the heuristic in every seat.

const (
	oracleOpeningLeylineOfSanctity = "492e0e6c-8c27-4376-938b-f8a8b6205810"
	oracleOpeningGemstoneCaverns   = "c0adbddc-b070-4c5f-afe0-0474c72a9251"
)

func TestHeuristicSeatsAnswerTheOpeningHandOffers(t *testing.T) {
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		if _, err := g.AddPlayer(fmt.Sprintf("Bot%d", i), myriadDeck(uuid.Nil, 8, 5)); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(3, 4))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Seat 0 (the starting player) holds a Leyline; seat 1 holds the
	// Caverns and two cards to pitch to it.
	hand := func(p *game.Player, cards ...game.Card) []uuid.UUID {
		p.Hand.Cards = nil
		var ids []uuid.UUID
		for _, c := range cards {
			c.InstanceID = uuid.New()
			c.Owner, c.Controller = p.ID, p.ID
			c.KnownBy = map[uuid.UUID]bool{p.ID: true}
			p.Hand.PushTop(c)
			ids = append(ids, c.InstanceID)
		}
		return ids
	}
	leyline := hand(g.Seats[0], game.Card{Name: "Leyline of Sanctity", TypeLine: "Enchantment", ManaCost: "{2}{W}{W}", OracleID: oracleOpeningLeylineOfSanctity})[0]
	cave := hand(g.Seats[1],
		game.Card{Name: "Gemstone Caverns", TypeLine: "Legendary Land", OracleID: oracleOpeningGemstoneCaverns},
		game.Card{Name: "Spare One", TypeLine: "Sorcery", ManaCost: "{1}{R}"},
		game.Card{Name: "Spare Two", TypeLine: "Sorcery", ManaCost: "{2}{R}"},
	)[0]
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}

	pol := heuristic.New()
	for step := 0; step < 20 && len(g.PendingChoices) > 0; step++ {
		var actor uuid.UUID
		var moves []legal.Move
		for _, s := range g.Seats {
			if m := legal.EnumerateFor(g, s.ID); len(m) > 0 {
				actor, moves = s.ID, m
				break
			}
		}
		if actor == uuid.Nil {
			t.Fatalf("WEDGE at step %d: a prompt is open and no seat has a move: %+v", step, g.PendingChoices)
		}
		d, err := pol.Decide(context.Background(), aiseat.Input{View: protocol.ViewOfGameFor(g, actor.String()), Seat: actor, Moves: moves})
		if err != nil {
			t.Fatalf("Decide: %v", err)
		}
		if d.Index == aiseat.Decline {
			t.Fatalf("the policy declined to answer %v", moves)
		}
		m := moves[d.Index]
		if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: actor, Caller: actor, Params: m.Params}); err != nil {
			t.Fatalf("step %d: %q rejected: %v", step, m.Label, err)
		}
	}
	if len(g.PendingChoices) != 0 {
		t.Fatalf("the opening never finished: %+v", g.PendingChoices)
	}
	if !g.Battlefield.Contains(leyline) {
		t.Error("the heuristic declined a free Leyline")
	}
	if !g.Battlefield.Contains(cave) {
		t.Fatal("the heuristic declined a free Gemstone Caverns")
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == cave && c.Counters["luck"] != 1 {
			t.Errorf("luck counters = %v, want 1", c.Counters)
		}
	}
	if g.Exile.Size() != 1 || g.Seats[1].Hand.Size() != 1 {
		t.Errorf("exile %d, hand %d: want one card exiled and one left", g.Exile.Size(), g.Seats[1].Hand.Size())
	}
}

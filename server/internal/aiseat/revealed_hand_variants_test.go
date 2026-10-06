package aiseat_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #2115: the revealed-hand pick's variants reach the bot as legal
// moves it can answer. An optional exile with the graveyard open is
// offered "choose nothing" (always legal) and one move per candidate,
// hand and graveyard alike, never a card the pick does not allow; the
// heuristic's answer is one of them and the engine accepts it.

// rhvPickMoves is every resolve_choice move, with its card IDs.
func rhvPickMoves(t *testing.T, moves []legal.Move) map[int][]string {
	t.Helper()
	out := map[int][]string{}
	for i, m := range moves {
		var p struct {
			CardIDs []string `json:"card_ids"`
		}
		if m.Type != "resolve_choice" || json.Unmarshal(m.Params, &p) != nil {
			continue
		}
		out[i] = p.CardIDs
	}
	return out
}

func TestTheBotAnswersAnOptionalExilePickWithTheGraveyardOpen(t *testing.T) {
	g := newSettledTable(t, 21)
	bot := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	land, spell, grave := uuid.New(), uuid.New(), uuid.New()
	g.WithWriteLock(func() {
		victim.Hand.Cards = nil
		victim.Hand.PushTop(game.Card{InstanceID: land, Name: "Gaea's Cradle", TypeLine: "Legendary Land", Owner: victim.ID, Controller: victim.ID})
		victim.Hand.PushTop(game.Card{InstanceID: spell, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}", Owner: victim.ID, Controller: victim.ID})
		victim.Graveyard.Cards = nil
		victim.Graveyard.PushTop(game.Card{InstanceID: grave, Name: "Giant Growth", TypeLine: "Instant", ManaCost: "{G}", Owner: victim.ID, Controller: victim.ID})
		if _, err := g.RevealedHandPickForEffect(game.RevealedHandDiscard{
			Chooser: bot.ID, FromPlayer: victim.ID, Count: 1, Reason: "Test",
			Filter: func(c game.Card) bool { return !c.IsLand() },
			Label:  "nonland card", Destination: game.PickExile, Optional: true, FromGraveyard: true,
		}); err != nil {
			t.Fatalf("RevealedHandPickForEffect: %v", err)
		}
	})

	moves := legal.EnumerateFor(g, bot.ID)
	picks := rhvPickMoves(t, moves)
	empty, named := 0, map[string]bool{}
	for i, ids := range picks {
		if len(ids) == 0 {
			empty++
			if !moves[i].AlwaysLegal {
				t.Error("choosing nothing is not marked always legal")
			}
		}
		for _, id := range ids {
			named[id] = true
		}
	}
	if empty != 1 || len(picks) != 3 || !named[spell.String()] || !named[grave.String()] || named[land.String()] {
		t.Fatalf("offered %v, want choose-nothing plus the creature and the graveyard instant, never the land", picks)
	}

	in := aiseat.Input{View: protocol.ViewOfGameFor(g, bot.ID.String()), Seat: bot.ID, Moves: moves}
	d, err := heuristic.New().Decide(context.Background(), in)
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if _, ok := picks[d.Index]; !ok {
		t.Fatalf("the heuristic answered with %+v, not a pick", d)
	}
	if len(picks[d.Index]) == 0 {
		t.Error("the heuristic chose nothing from an opponent's hand with a card worth taking")
	}
	chosen := moves[d.Index]
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(chosen.Type), Player: bot.ID, Caller: bot.ID, Params: chosen.Params,
	}); err != nil {
		t.Fatalf("the engine refused the bot's pick %q: %v", chosen.Label, err)
	}
	if !g.Exile.Contains(spell) && !g.Exile.Contains(grave) {
		t.Error("the bot's pick exiled nothing")
	}
}

// A bot that must pick from its own revealed hand, and may choose
// nothing, chooses nothing: giving up its own card is worth less than
// keeping it.
func TestTheBotChoosesNothingFromItsOwnHandWhenItMay(t *testing.T) {
	g := newSettledTable(t, 21)
	bot := g.Seats[g.Turn.ActiveSeat]
	spell := uuid.New()
	g.WithWriteLock(func() {
		bot.Hand.Cards = nil
		bot.Hand.PushTop(game.Card{InstanceID: spell, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}", Owner: bot.ID, Controller: bot.ID})
		if _, err := g.RevealedHandPickForEffect(game.RevealedHandDiscard{
			Chooser: bot.ID, FromPlayer: bot.ID, Count: 1, Reason: "Test", Optional: true,
		}); err != nil {
			t.Fatalf("RevealedHandPickForEffect: %v", err)
		}
	})
	moves := legal.EnumerateFor(g, bot.ID)
	picks := rhvPickMoves(t, moves)
	in := aiseat.Input{View: protocol.ViewOfGameFor(g, bot.ID.String()), Seat: bot.ID, Moves: moves}
	d, err := heuristic.New().Decide(context.Background(), in)
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if ids, ok := picks[d.Index]; !ok || len(ids) != 0 {
		t.Errorf("the heuristic answered %q, want choose nothing", moves[d.Index].Label)
	}
}

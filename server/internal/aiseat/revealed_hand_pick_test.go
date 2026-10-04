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

// ADR 0116 §8 (#2078): the bot picks only matching cards. A bot that
// cast Thoughtseize is offered no move naming the land in the revealed
// hand, the heuristic's answer is one of the nonland cards, and the
// engine accepts it.
func TestTheBotPicksOnlyAnEligibleCardFromARevealedHand(t *testing.T) {
	g := newSettledTable(t, 21)
	bot := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]

	land := uuid.New()
	var spells []uuid.UUID
	g.WithWriteLock(func() {
		victim.Hand.Cards = nil
		// A land the bot would happily take if it could: several of
		// them would be the "best card" by any count of value.
		victim.Hand.PushTop(game.Card{InstanceID: land, Name: "Gaea's Cradle", TypeLine: "Legendary Land", Owner: victim.ID, Controller: victim.ID})
		for _, name := range []string{"Llanowar Elves", "Giant Growth"} {
			id := uuid.New()
			tl := "Creature — Elf Druid"
			if name == "Giant Growth" {
				tl = "Instant"
			}
			victim.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: tl, ManaCost: "{G}", Owner: victim.ID, Controller: victim.ID})
			spells = append(spells, id)
		}
		g.QueueDiscardFromRevealedHand(game.RevealedHandDiscard{
			Chooser: bot.ID, FromPlayer: victim.ID, Count: 1, Reason: "Thoughtseize",
			Filter: func(c game.Card) bool { return !c.IsLand() },
			Label:  "nonland card",
		})
	})

	moves := legal.EnumerateFor(g, bot.ID)
	answers := 0
	for _, m := range moves {
		var p struct {
			CardIDs []string `json:"card_ids"`
		}
		if m.Type != "resolve_choice" || json.Unmarshal(m.Params, &p) != nil {
			continue
		}
		answers++
		for _, id := range p.CardIDs {
			if id == land.String() {
				t.Errorf("the enumerator offered the land: %s", m.Label)
			}
		}
	}
	if answers != len(spells) {
		t.Fatalf("offered %d answers, want one per nonland card (%d)", answers, len(spells))
	}

	in := aiseat.Input{View: protocol.ViewOfGameFor(g, bot.ID.String()), Seat: bot.ID, Moves: moves}
	d, err := heuristic.New().Decide(context.Background(), in)
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if d.Index < 0 || d.Index >= len(moves) {
		t.Fatalf("the heuristic declined a mandatory pick: %+v", d)
	}
	chosen := moves[d.Index]
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(chosen.Type), Player: bot.ID, Caller: bot.ID, Params: chosen.Params,
	}); err != nil {
		t.Fatalf("the engine refused the bot's pick %q: %v", chosen.Label, err)
	}
	discarded := 0
	for _, id := range spells {
		if victim.Graveyard.Contains(id) {
			discarded++
		}
	}
	if discarded != 1 || !victim.Hand.Contains(land) {
		t.Errorf("discarded %d nonland cards; the land must stay in hand", discarded)
	}
}

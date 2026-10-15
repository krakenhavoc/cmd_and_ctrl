package aiseat_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// exile_card_cast_cost_test.go — #3536 through the real engine: the bot
// casts Abhorrent Oculus out of a seven-card graveyard, paying its
// "exile six cards from your graveyard" additional cost.
//
// Not behind AISEAT_GAME_TESTS: one move list, one decision, one dispatch.

func TestHeuristicCastsAbhorrentOculusFromASevenCardGraveyard(t *testing.T) {
	g := newSettledTable(t, 3536)
	advanceToStep(t, g, game.StepPrecombatMain)
	me := g.Seats[g.Turn.ActiveSeat]
	everyone := seatIDs(g)

	var oculus uuid.UUID
	var fuel []uuid.UUID
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		c := game.NewCard("Abhorrent Oculus", me.ID)
		c.TypeLine = "Creature — Eye"
		c.ManaCost = "{2}{U}"
		c.Power, c.Toughness = 5, 5
		c.OracleID = "b1a0e129-27f0-4a09-b00d-497e18777306"
		c.Controller = me.ID
		c.AddKnowersAll(everyone)
		me.Hand.PushTop(c)
		oculus = c.InstanceID
		for range 7 {
			f := game.NewCard("Spent Spell", me.ID)
			f.TypeLine = "Instant"
			f.Controller = me.ID
			f.AddKnowersAll(everyone)
			me.Graveyard.PushTop(f)
			fuel = append(fuel, f.InstanceID)
		}
	})
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{U}"); err != nil {
		t.Fatal(err)
	}

	moves := legal.EnumerateFor(g, me.ID)
	in := aiseat.Input{View: protocol.ViewOfGameFor(g, me.ID.String()), Seat: me.ID, Moves: moves}
	d, err := heuristic.New().Decide(context.Background(), in)
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	if d.Index < 0 || d.Index >= len(moves) {
		t.Fatalf("the bot made no move (index %d of %d)", d.Index, len(moves))
	}
	mv := moves[d.Index]
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(mv.Type), Player: mv.Player, Caller: me.ID, Params: mv.Params}); err != nil {
		t.Fatalf("dispatch %q: %v", mv.Label, err)
	}
	if !g.Stack.Contains(oculus) {
		t.Fatalf("the bot chose %q; want it to cast Abhorrent Oculus", mv.Label)
	}
	exiled := 0
	for _, id := range fuel {
		if g.Exile.Contains(id) {
			exiled++
		}
	}
	if exiled != 6 || me.Graveyard.Size() != 1 {
		t.Errorf("%d fuel cards exiled and %d left in the graveyard, want 6 and 1", exiled, me.Graveyard.Size())
	}
}

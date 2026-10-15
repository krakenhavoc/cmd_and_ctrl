package aiseat_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// extra_turn_bot_test.go — ADR 0188 Decision 10, #3566 through the real
// card, the real enumerator and the real heuristic: a bot holding Time
// Warp at an even board casts it at ITSELF, the turn lands on the queue,
// and the view then names that turn as the next one. Not behind
// AISEAT_GAME_TESTS: one cast, one resolution.

const botTimeWarpOracle = "dbd6a94b-62ff-4a10-9d52-bdd90b26e425"

func newSettledFourSeatTable(t *testing.T, seed uint64) *game.Game {
	t.Helper()
	g := newRoom(t, 4, seed).Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

func TestHeuristicSeatCastsTimeWarpAtItself(t *testing.T) {
	g := newSettledFourSeatTable(t, 3566)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)
	me.Hand.Cards = nil
	for i := 0; i < 5; i++ {
		seedBattlefield(g, me, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	}
	warp := game.Card{
		InstanceID: uuid.New(), Name: "Time Warp", OracleID: botTimeWarpOracle,
		TypeLine: "Sorcery", ManaCost: "{3}{U}{U}", Colors: []string{"U"},
		Owner: me.ID, Controller: me.ID,
	}
	warp.AddKnowersAll(seatIDs(g))
	me.Hand.PushTop(warp)

	pol := heuristic.New()
	moves := legal.EnumerateFor(g, me.ID)
	in := aiseat.Input{View: protocol.ViewOfGameFor(g, me.ID.String()), Seat: me.ID, Moves: moves}
	d, err := pol.Decide(context.Background(), in)
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if d.Index < 0 || d.Index >= len(moves) {
		t.Fatalf("decision %d out of range of %d moves", d.Index, len(moves))
	}
	mv := moves[d.Index]
	if mv.Type != legal.TypeCastSpell || mv.Source != warp.InstanceID {
		t.Fatalf("the bot chose %q; with Time Warp and five Islands it should cast it; moves %v", mv.Label, attackLabels(moves))
	}
	if !strings.Contains(string(mv.Params), me.ID.String()) {
		t.Fatalf("the bot aimed Time Warp at someone else: %s", mv.Params)
	}
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(mv.Type), Player: mv.Player, Caller: me.ID, Params: mv.Params,
	}); err != nil {
		t.Fatalf("the engine refused the enumerated cast %q: %v", mv.Label, err)
	}
	for i := 0; i < 16 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	queued := g.ExtraTurnsQueued()
	if len(queued) != 1 || g.Seats[queued[0].Seat].ID != me.ID {
		t.Fatalf("extra turns queued = %+v, want one for the bot's seat", queued)
	}

	// The view now names that turn as the next one, to every viewer.
	for _, viewer := range g.Seats {
		v := protocol.ViewOfGameFor(g, viewer.ID.String())
		if v.Turn.NextSeat == nil || *v.Turn.NextSeat != queued[0].Seat {
			t.Errorf("viewer %s sees next_seat %v, want %d", viewer.Name, v.Turn.NextSeat, queued[0].Seat)
		}
	}
}

// TestNextSeatInASecondSeatsExtraTurn is the live bug end to end on a
// real four-seat game: seat 0 gives seat 2 an extra turn; in that turn the
// view names seat 1, and the next turn that really begins is seat 1's.
func TestNextSeatInASecondSeatsExtraTurn(t *testing.T) {
	g := newSettledFourSeatTable(t, 3567)
	if g.Turn.ActiveSeat != 0 {
		t.Skipf("the seed starts seat %d, not seat 0", g.Turn.ActiveSeat)
	}
	g.WithWriteLock(func() { g.TakeExtraTurnsForEffect(g.Seats[2].ID, uuid.Nil, 1) })
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.Turn.ActiveSeat != 2 || !g.Turn.Extra {
		t.Fatalf("turn is seat %d extra=%v, want seat 2's extra turn", g.Turn.ActiveSeat, g.Turn.Extra)
	}
	for _, viewer := range g.Seats {
		v := protocol.ViewOfGameFor(g, viewer.ID.String())
		if v.Turn.NextSeat == nil || *v.Turn.NextSeat != 1 {
			t.Errorf("viewer %s sees next_seat %v in seat 2's extra turn, want 1", viewer.Name, v.Turn.NextSeat)
		}
	}
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	if g.Turn.ActiveSeat != 1 {
		t.Errorf("after the extra turn seat %d began, want seat 1", g.Turn.ActiveSeat)
	}
}

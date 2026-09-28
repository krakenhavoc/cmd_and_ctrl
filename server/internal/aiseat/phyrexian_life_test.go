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

// phyrexian_life_test.go — #1677, whole-engine: the real enumerator,
// the real heuristic and the real catalog Gut Shot ({R/P}: 1 damage to
// any target). With a Mountain the bot pays the {R/P} with mana even
// though the life payment is on offer; with no red source it pays 2
// life, and the engine charges exactly that.
//
// Not behind AISEAT_GAME_TESTS: it plays no whole game.

const botGutShotOracle = "9afb3b6e-4909-4efa-aa79-81c0229411c9"

// gutShotTable parks a two-seat game on the active seat's main phase
// holding a Gut Shot, with `mountains` Mountains, at `life`, facing an
// opponent's 1/1 the ping kills.
func gutShotTable(t *testing.T, mountains, life int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newSettledTable(t, 1677)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToStep(t, g, game.StepPrecombatMain)
	me.Hand.Cards = nil
	me.Life = life

	push := func(owner *game.Player, name, typeLine string, power, tough int) {
		c := game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
			Power: power, Toughness: tough, Owner: owner.ID, Controller: owner.ID,
		}
		c.AddKnowersAll(seatIDs(g))
		g.Battlefield.PushTop(c)
	}
	push(them, "Pestering Sprite", "Creature — Faerie", 1, 1)
	for i := 0; i < mountains; i++ {
		push(me, "Mountain", "Basic Land — Mountain", 0, 0)
	}
	shot := game.Card{
		InstanceID: uuid.New(), Name: "Gut Shot", TypeLine: "Instant",
		ManaCost: "{R/P}", OracleID: botGutShotOracle, Layout: "normal",
		Owner: me.ID, Controller: me.ID,
		KnownBy: map[uuid.UUID]bool{me.ID: true},
	}
	me.Hand.PushTop(shot)
	return g, me, shot.InstanceID
}

func castPhyrexianLife(t *testing.T, m legal.Move) int {
	t.Helper()
	var p struct {
		PhyrexianLife int `json:"phyrexian_life"`
	}
	if err := json.Unmarshal(m.Params, &p); err != nil {
		t.Fatalf("params %s: %v", string(m.Params), err)
	}
	return p.PhyrexianLife
}

// decideGutShot runs the heuristic over the Gut Shot casts plus the
// pass — the commander in the command zone is also castable off three
// Mountains, and this test is about how the Gut Shot is paid, not
// which spell the bot likes best.
func decideGutShot(t *testing.T, g *game.Game, me *game.Player, shot uuid.UUID) (legal.Move, map[int]bool) {
	t.Helper()
	var moves []legal.Move
	offered := map[int]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		switch {
		case m.Kind == legal.KindPass:
			moves = append(moves, m)
		case m.Kind == legal.KindCast && m.Source == shot:
			moves = append(moves, m)
			offered[castPhyrexianLife(t, m)] = true
		}
	}
	d, err := heuristic.New().Decide(context.Background(), aiseat.Input{
		View:  protocol.ViewOfGameFor(g, me.ID.String()),
		Seat:  me.ID,
		Moves: moves,
	})
	if err != nil {
		t.Fatalf("policy: %v", err)
	}
	if d.Index < 0 || d.Index >= len(moves) {
		t.Fatalf("policy picked index %d of %d", d.Index, len(moves))
	}
	return moves[d.Index], offered
}

func dispatchMove(t *testing.T, g *game.Game, m legal.Move) {
	t.Helper()
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(m.Type), Player: m.Player, Caller: m.Player, Params: m.Params,
	}); err != nil {
		t.Fatalf("engine refused the bot's pick %q: %v", m.Label, err)
	}
}

// TestBotPaysGutShotWithManaWhenItHasIt — both payments are offered,
// and the heuristic takes the mana one.
func TestBotPaysGutShotWithManaWhenItHasIt(t *testing.T) {
	g, me, shot := gutShotTable(t, 3, 40)
	pick, offered := decideGutShot(t, g, me, shot)
	if !offered[0] || !offered[1] {
		t.Fatalf("offered phyrexian_life %v, want both the mana (0) and the life (1) payment", offered)
	}
	if pick.Kind != legal.KindCast {
		t.Fatalf("the bot passed instead of pinging a 1/1 for {R}: %q", pick.Label)
	}
	if n := castPhyrexianLife(t, pick); n != 0 {
		t.Errorf("the bot paid %d symbol(s) with life holding three Mountains: %q", n, pick.Label)
	}
	dispatchMove(t, g, pick)
	if me.Life != 40 {
		t.Errorf("life after a mana-paid Gut Shot = %d, want 40", me.Life)
	}
}

// TestBotPaysGutShotWithLifeWhenItMustAndCan — no red source: the
// life payment is the only cast, the bot takes it at 40, and the engine
// takes 2 life for it.
func TestBotPaysGutShotWithLifeWhenItMustAndCan(t *testing.T) {
	g, me, shot := gutShotTable(t, 0, 40)
	pick, offered := decideGutShot(t, g, me, shot)
	if offered[0] || !offered[1] {
		t.Fatalf("offered phyrexian_life %v, want only the life payment (1)", offered)
	}
	if pick.Kind != legal.KindCast || castPhyrexianLife(t, pick) != 1 {
		t.Fatalf("the bot did not pay 2 life for Gut Shot at 40: %q", pick.Label)
	}
	dispatchMove(t, g, pick)
	if me.Life != 38 {
		t.Errorf("life after a life-paid Gut Shot = %d, want 38", me.Life)
	}
}

// TestBotDoesNotPayPhyrexianLifeBelowTheFloor — same board at 11 life:
// paying 2 would leave 9, under the floor of 10, so the bot passes.
func TestBotDoesNotPayPhyrexianLifeBelowTheFloor(t *testing.T) {
	g, me, shot := gutShotTable(t, 0, 11)
	pick, offered := decideGutShot(t, g, me, shot)
	if !offered[1] {
		t.Fatalf("offered phyrexian_life %v, want the life payment on offer (11 life can pay 2)", offered)
	}
	if pick.Kind == legal.KindCast {
		t.Errorf("the bot paid Phyrexian life down to %d: %q", me.Life-2, pick.Label)
	}
}

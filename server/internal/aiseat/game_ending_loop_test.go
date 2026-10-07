package aiseat_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// game_ending_loop_test.go — #2450, ADR 0055's amendment of
// 2026-10-07, option B, whole-engine with the real catalog.
//
// Exquisite Blood with Marauding Blight-Priest is a real mandatory
// loop: each gain drains every opponent by 1, and each drain gains
// life again. It adds more triggers than it resolves, so option A
// never sees it drain. But every Blight-Priest trigger takes three
// life totals to a new low, and a life total makes only so many lows
// before CR 704.5a takes its player. So the loop keeps stepping until
// every opponent is gone and its controller wins (CR 104.2a).
//
// Not behind AISEAT_GAME_TESTS: one loop, well under a second.

const botBlightPriestOracle = "814b87fe-2a75-4ff2-8637-7e69e3fb285b"

func TestAGameEndingLoopStepsToAWin(t *testing.T) {
	const oppLife = 40
	g := newRoom(t, 4, 2451).Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)
	for _, p := range g.Seats {
		p.Hand.Cards = nil
		if p.ID != me.ID {
			p.Life = oppLife
		}
	}
	enterForTest(g, game.Card{Name: "Marauding Blight-Priest", OracleID: botBlightPriestOracle,
		TypeLine: "Creature — Vampire Cleric", Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	enterForTest(g, game.Card{Name: "Exquisite Blood", OracleID: botExquisiteBloodOracle,
		TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })

	pol := heuristic.New()
	for step := 0; step < 20000 && g.State == game.StateActive; step++ {
		var seat uuid.UUID
		g.ReadSnapshot(func() {
			for _, c := range g.PendingChoices {
				if c != nil {
					seat = c.Chooser
					return
				}
			}
			if h := g.Turn.PriorityHolder; h >= 0 && h < len(g.Seats) {
				seat = g.Seats[h].ID
			}
		})
		if seat == uuid.Nil {
			t.Fatalf("step %d: nobody to act in %s", step, g.Turn.Step)
		}
		moves := legal.EnumerateFor(g, seat)
		if len(moves) == 0 {
			t.Fatalf("step %d: %s was offered nothing", step, seat)
		}
		d, err := pol.Decide(context.Background(), aiseat.Input{
			View: protocol.ViewOfGameFor(g, seat.String()), Seat: seat, Moves: moves,
		})
		if err != nil || d.Index < 0 || d.Index >= len(moves) {
			t.Fatalf("step %d: policy: index %d, %v", step, d.Index, err)
		}
		mv := moves[d.Index]
		if mv.Kind == legal.KindPass && g.AutoPassSuspended() {
			t.Fatalf("step %d: the loop breaker stopped a loop that is ending the game (%+v); every bot seat would hold here — #2450",
				step, g.CurrentLoopNotice())
		}
		dispatchMove(t, g, mv)
	}
	if g.State == game.StateActive {
		t.Fatal("the drain never finished the game")
	}
	for _, p := range g.Seats {
		if p.ID != me.ID && !p.Eliminated {
			t.Errorf("%s is still in the game at %d life", p.Name, p.Life)
		}
	}
	if me.Eliminated {
		t.Error("the loop's controller lost")
	}
}

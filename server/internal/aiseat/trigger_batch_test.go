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

// trigger_batch_test.go — #2450, whole-engine: the real catalog Syr
// Konrad and Exquisite Blood, the real enumerator and the real
// heuristic at every seat.
//
// ADR 0126's runs 1 and 2 stopped three bot-only tables at the CR 732
// loop breaker. Seeds 15 and 21 were this: thirty creatures dying at
// once with Syr Konrad and Exquisite Blood out. Each Konrad trigger
// pings three opponents and each ping triggers Exquisite Blood, so
// Blood's key passed 25 resolutions with no decision in between. The
// bot took the shortcut's ten more, the second ask offered only
// "stop", and every seat held with Konrad triggers still waiting. That
// is a finite batch (CR 603.2c), not a loop, and ADR 0055's amendment
// of 2026-10-07 (option A) says a batch that drains the stack does not
// trip the breaker.
//
// Not behind AISEAT_GAME_TESTS: one batch, milliseconds.

const (
	botKonradOracle         = "14c3ff84-1e82-4606-a433-869fc52cc382"
	botExquisiteBloodOracle = "8f933fae-6c0c-42d7-a817-14760d8285cd"
)

// enterForTest puts c onto the battlefield and announces it, so the
// catalog's triggers see it as they would a permanent that entered.
func enterForTest(g *game.Game, c game.Card) uuid.UUID {
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	c.AddKnowersAll(seatIDs(g))
	g.Battlefield.PushTop(c)
	id := c.InstanceID
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	return id
}

func TestAKonradBatchDoesNotTripTheLoopBreaker(t *testing.T) {
	const deaths = 30
	g := newRoom(t, 4, 2450).Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	me := g.Seats[g.Turn.ActiveSeat]
	victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToStep(t, g, game.StepPrecombatMain)
	for _, p := range g.Seats {
		p.Hand.Cards = nil
		if p.ID != me.ID {
			p.Life = 200
		}
	}
	enterForTest(g, game.Card{Name: "Syr Konrad, the Grim", OracleID: botKonradOracle,
		TypeLine: "Legendary Creature — Human Knight", Power: 5, Toughness: 4, Owner: me.ID, Controller: me.ID})
	enterForTest(g, game.Card{Name: "Exquisite Blood", OracleID: botExquisiteBloodOracle,
		TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID})
	var plants []uuid.UUID
	for i := 0; i < deaths; i++ {
		plants = append(plants, enterForTest(g, game.Card{Name: "Plant", TypeLine: "Creature — Plant",
			Power: 0, Toughness: 1, Owner: victim.ID, Controller: victim.ID}))
	}
	lifeBefore := me.Life
	turn := g.Turn.Round
	g.WithWriteLock(func() {
		if n := g.DestroyPermanentsForEffect(plants); n != deaths {
			t.Fatalf("destroyed %d, want %d", n, deaths)
		}
	})

	// Every seat plays the heuristic until the turn ends. A pass while
	// automatic passing is suspended is where a bot runner would hold
	// (runner.go), and that is the stall.
	pol := heuristic.New()
	for step := 0; step < 2000 && g.Turn.Round == turn && g.Turn.Step != game.StepEnd; step++ {
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
			t.Fatalf("step %d: the loop breaker fired on a batch of %d deaths (%+v); every bot seat would hold here — #2450",
				step, deaths, g.CurrentLoopNotice())
		}
		dispatchMove(t, g, mv)
	}
	if g.Turn.Round == turn && g.Turn.Step != game.StepEnd {
		t.Fatalf("the turn did not finish (at %s)", g.Turn.Step)
	}
	// The whole batch resolved: every Konrad trigger pinged every
	// opponent, and every ping fed Exquisite Blood.
	for _, p := range g.Seats {
		if p.ID != me.ID && p.Life != 200-deaths {
			t.Errorf("%s at %d life, want %d: the batch did not finish", p.Name, p.Life, 200-deaths)
		}
	}
	if want := lifeBefore + deaths*(len(g.Seats)-1); me.Life != want {
		t.Errorf("Konrad's controller at %d life, want %d", me.Life, want)
	}
}

package legal_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// power_floor_test.go — the #544 half of #4678. Technodrome "can't
// attack or block unless its power is 6 or greater": the enumerator
// offers its attacks and blocks exactly when the engine accepts them,
// reading its power after layers and counters.

const oracleTechnodrome = "51c0b46c-c379-420a-a4a4-d18724c759aa"

// technodromeFor puts a Technodrome under p, entered (so the layer pass
// reaches its static) but not summoning sick.
func technodromeFor(g *game.Game, p *game.Player) uuid.UUID {
	id := enteredCard(g, p, game.Card{
		Name: "Technodrome", TypeLine: "Artifact Creature — Construct", OracleID: oracleTechnodrome,
		Power: 3, Toughness: 3, Keywords: []string{"reach", "trample"},
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if c := &g.Battlefield.Cards[i]; c.InstanceID == id {
				c.SummonedThisTurn = false
			}
		}
	})
	return id
}

func addCounters(t *testing.T, g *game.Game, id uuid.UUID, kind string, n int) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.AddCounterForEffect(id, kind, n) })
	if err != nil {
		t.Fatalf("AddCounterForEffect: %v", err)
	}
}

func TestTechnodromeAttacksAreOfferedOnlyAtPowerSix(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	drome := technodromeFor(g, me)
	free := freshCreature(g, me, "Free Bear")
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	if n := attacksBy(moves, drome); n != 0 {
		t.Errorf("a 3/3 Technodrome was offered %d attacks: %v", n, labels(moves))
	}
	if attacksBy(moves, free) == 0 {
		t.Fatalf("no attack offered for the unrestricted creature: %v", labels(moves))
	}
	if err := g.Clone().DeclareAttacker(drome, them.ID); !errors.Is(err, game.ErrCantAttack) {
		t.Errorf("engine accepted the withheld attack: %v", err)
	}

	addCounters(t, g, drome, game.CounterPlusOne, 3)
	moves = legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	if attacksBy(moves, drome) == 0 {
		t.Fatalf("a 6/6 Technodrome was offered no attack: %v", labels(moves))
	}

	addCounters(t, g, drome, game.CounterMinusOne, 1)
	moves = legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	if n := attacksBy(moves, drome); n != 0 {
		t.Errorf("a 5/5 Technodrome was offered %d attacks: %v", n, labels(moves))
	}
}

func TestTechnodromeBlocksAreOfferedOnlyAtPowerSix(t *testing.T) {
	for _, tc := range []struct {
		plus  int
		block bool
	}{{0, false}, {3, true}} {
		g := newTable(t)
		me := g.Seats[g.Turn.ActiveSeat]
		them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		clearHand(me)
		clearHand(them)
		attacker := freshCreature(g, me, "Attacking Bear")
		drome := technodromeFor(g, them)
		if tc.plus > 0 {
			addCounters(t, g, drome, game.CounterPlusOne, tc.plus)
		}
		advanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(attacker, them.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		advanceTo(t, g, game.StepDeclareBlockers)

		moves := legal.EnumerateFor(g, them.ID)
		dispatchAll(t, g, them.ID, moves)
		offered := blocksBy(moves, drome) > 0
		if offered != tc.block {
			t.Errorf("with %d +1/+1 counters: block offered %v, want %v: %v", tc.plus, offered, tc.block, labels(moves))
		}
		err := g.Clone().DeclareBlocker(drome, attacker)
		if tc.block && err != nil {
			t.Errorf("with %d +1/+1 counters the engine refused the block: %v", tc.plus, err)
		}
		if !tc.block && !errors.Is(err, game.ErrIllegalBlock) {
			t.Errorf("with %d +1/+1 counters the engine accepted the withheld block: %v", tc.plus, err)
		}
	}
}

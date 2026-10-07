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

// equip_loop_test.go — #2449, whole-engine: the real catalog Lightning
// Greaves, the real enumerator and the real heuristic.
//
// The catalog soak at seed 2026100600 (run 37504334671, on ADR 0126 PR
// 4) stalled at turn 5: a heuristic seat moved its Lightning Greaves
// (Equip {0}) between Ray Fillet, Wave Warrior and Covetous Castaway
// until the CR 732 breaker named the ability, and the runner then held
// on it (#810) with the policy still choosing it, so nothing committed
// again. A seed fixes the decks dealt, not the play, and the catalog has
// grown since, so the seed no longer deals that board; this test builds
// it instead.
//
// Not behind AISEAT_GAME_TESTS: one main phase, milliseconds.

const (
	botGreavesOracle   = "ca204b66-8d0c-431a-8d34-282f7c2d17da"
	botRayFilletOracle = "206913aa-4d73-4465-8311-8054a22c0743"
	botCastawayOracle  = "8f3e6554-eb8a-4096-81a3-2411186d9cb4"
)

// TestHeuristicSeatEquipsGreavesOnceAndMovesOn: the bot's main phase
// with an unattached Greaves and the two creatures the soak had. The
// bot plays its main phase out, the opponent passing back each time the
// stack needs it. The first equip is real and the bot may take it; a
// second move between its own creatures buys nothing, and the main
// phase has to end.
func TestHeuristicSeatEquipsGreavesOnceAndMovesOn(t *testing.T) {
	g := newSettledTable(t, 2449)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToStep(t, g, game.StepPrecombatMain)
	me.Hand.Cards = nil

	push := func(c game.Card) uuid.UUID {
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = me.ID, me.ID
		c.AddKnowersAll(seatIDs(g))
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	push(game.Card{Name: "Ray Fillet, Wave Warrior", OracleID: botRayFilletOracle,
		TypeLine: "Legendary Creature — Fish Mutant", ManaCost: "{2}{U}", Power: 0, Toughness: 2})
	push(game.Card{Name: "Covetous Castaway", OracleID: botCastawayOracle,
		TypeLine: "Creature — Human", ManaCost: "{1}{U}", Power: 1, Toughness: 3})
	greaves := push(game.Card{Name: "Lightning Greaves", OracleID: botGreavesOracle,
		TypeLine: "Artifact — Equipment", ManaCost: "{2}"})

	pol := heuristic.New()
	var equips []string
	for step := 0; step < 120; step++ {
		if g.Turn.Step != game.StepPrecombatMain || owesAnyChoice(g) {
			break
		}
		if g.Turn.PriorityHolder < 0 || g.Turn.PriorityHolder >= len(g.Seats) {
			t.Fatalf("step %d: nobody holds priority in %s", step, g.Turn.Step)
		}
		seat := g.Seats[g.Turn.PriorityHolder].ID
		moves := legal.EnumerateFor(g, seat)
		if len(moves) == 0 {
			t.Fatalf("step %d: %s was offered nothing", step, seat)
		}
		var mv legal.Move
		if seat != me.ID {
			pi := aiseat.PassIndex(moves)
			if pi < 0 {
				t.Fatalf("step %d: the opponent has no pass on offer", step)
			}
			mv = moves[pi]
		} else {
			d, err := pol.Decide(context.Background(), aiseat.Input{
				View: protocol.ViewOfGameFor(g, seat.String()), Seat: seat, Moves: moves,
			})
			if err != nil {
				t.Fatalf("step %d: policy: %v", step, err)
			}
			if d.Index < 0 || d.Index >= len(moves) {
				t.Fatalf("step %d: policy picked index %d of %d", step, d.Index, len(moves))
			}
			mv = moves[d.Index]
			if mv.Kind == legal.KindActivate && mv.Source == greaves {
				equips = append(equips, mv.Label)
			}
		}
		dispatchMove(t, g, mv)
	}
	if g.Turn.Step == game.StepPrecombatMain {
		t.Fatalf("the bot never left its main phase; it equipped the Greaves %d times (%v) — #2449", len(equips), equips)
	}
	// One equip onto the better host, and at most one move after it
	// when a host proves better: never back and forth.
	if len(equips) > 2 {
		t.Errorf("the bot equipped the Greaves %d times in one main phase: %v (#2449)", len(equips), equips)
	}
	if n := g.CurrentLoopNotice(); n != nil {
		t.Errorf("the CR 732 breaker fired: %+v", n)
	}
}

// owesAnyChoice reports whether any seat owes a prompt — the driver
// above plays priority only.
func owesAnyChoice(g *game.Game) bool {
	n := 0
	g.ReadSnapshot(func() { n = owedChoicesLocked(g) })
	return n > 0
}

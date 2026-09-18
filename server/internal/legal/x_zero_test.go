package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// x_zero_test.go — #810, the enumerator's rule that a move whose whole
// effect is X is not offered at X=0 (internal/legal/x.go, CR 732.2a).
//
// The first test is the issue's own probe, in the shape the issue
// reported it: Soothsaying on the battlefield, no mana, precombat
// main. `EnumerateFor` offered `activate` at `x_value: 0`, and offered
// it again the moment it resolved, five times out of five — which on a
// bot table meant 79,519 applied actions in five minutes and turn 18.

const oracleTheGooseMother = "de595f1b-3f7d-45e0-a31b-ed23e5d1ee48"

// The probe, both halves: with no mana Soothsaying's "{X}: Look at
// the top X cards" is not a move at all, and with mana it is a move at
// X ≥ 1 and never at X=0.
func TestSoothsayingIsNotOfferedAtXZero(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	sooth := battlefieldCard(g, active, game.Card{
		Name: "Soothsaying", TypeLine: "Enchantment", ManaCost: "{U}", OracleID: oracleSoothsaying,
	})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if acts := activationsOf(moves, sooth); len(acts) != 0 {
		t.Fatalf("no mana, so looking at the top 0 cards is the only announcement and it is not a move: %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)

	// Three Islands: the {X} half is affordable at X=3 and the
	// {3}{U}{U} shuffle is not, so there is exactly one activation and
	// its X is the largest the seat can pay.
	mana(g, active, 3)
	moves = legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, sooth)
	if len(acts) != 1 {
		t.Fatalf("three mana buys the {X} half once, got %d: %v", len(acts), labels(acts))
	}
	if x := xValueOf(t, acts[0]); x < 1 {
		t.Errorf("Soothsaying offered at X=%d; the minimum announcement worth offering is 1", x)
	}
	dispatchAll(t, g, active.ID, moves)
}

// The same board, walked the way the issue's probe walked it: dispatch
// the offer and enumerate again. Before #810 the same X=0 activation
// came back every iteration with the turn no further on; now the mana
// is spent by the first activation and the second enumeration has no
// {X} activation to offer, so the sequence terminates on its own.
func TestAnXActivationDoesNotComeBackForever(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	sooth := battlefieldCard(g, active, game.Card{
		Name: "Soothsaying", TypeLine: "Enchantment", ManaCost: "{U}", OracleID: oracleSoothsaying,
	})
	mana(g, active, 2)
	advanceTo(t, g, game.StepPrecombatMain)

	seen := 0
	for i := 0; i < 5; i++ {
		moves := legal.EnumerateFor(g, active.ID)
		acts := activationsOf(moves, sooth)
		if len(acts) == 0 {
			break
		}
		seen++
		for _, m := range acts {
			if x := xValueOf(t, m); x == 0 {
				t.Fatalf("iteration %d offered a free X=0 activation — the loop #810 reported", i)
			}
		}
		dispatchOne(t, g, active.ID, acts[0])
	}
	if seen == 0 {
		t.Fatal("two mana should buy at least one activation")
	}
	if seen > 2 {
		t.Errorf("the {X} activation was offered %d times off two mana — it is not paying for itself", seen)
	}
}

// The rider half of the rule. The Goose Mother is {X}{G}{U} for a 2/2
// flier with an attack trigger, so X=0 leaves a real permanent behind
// and is a real (if small) play. It declares no XMatters and the
// enumerator still offers it — but only because two mana is all there
// is: the search takes the largest affordable X, so a third land moves
// the offer to X=1 on its own.
func TestACastWithAFixedRiderIsStillOfferedAtXZero(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	goose := handCard(active, game.Card{
		Name: "The Goose Mother", TypeLine: "Legendary Creature — Bird Hydra",
		ManaCost: "{X}{G}{U}", Power: 2, Toughness: 2, OracleID: oracleTheGooseMother,
	})
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, goose)
	if len(casts) != 1 {
		t.Fatalf("The Goose Mother off a Forest and an Island: %d casts, want 1: %v", len(casts), labels(moves))
	}
	if x := xValueOf(t, casts[0]); x != 0 {
		t.Errorf("X = %d, want 0 — two mana pays for the body and nothing else", x)
	}
	dispatchAll(t, g, active.ID, moves)

	battlefieldCard(g, active, basic("Forest", "Forest"))
	moves = legal.EnumerateFor(g, active.ID)
	casts = castMovesFor(moves, goose)
	if len(casts) != 1 {
		t.Fatalf("The Goose Mother off three lands: %d casts, want 1", len(casts))
	}
	if x := xValueOf(t, casts[0]); x != 1 {
		t.Errorf("X = %d, want 1 — the search takes the largest affordable X", x)
	}
	dispatchAll(t, g, active.ID, moves)
}

// dispatchOne applies one enumerated move to the live game, rather
// than to the clone dispatchAll uses — this file needs the board to
// actually move so the next enumeration answers a changed question.
func dispatchOne(t *testing.T, g *game.Game, seat uuid.UUID, m legal.Move) {
	t.Helper()
	err := actions.Dispatch(g, actions.Action{
		Type:   actions.Type(m.Type),
		Player: m.Player,
		Caller: seat,
		Params: m.Params,
	})
	if err != nil {
		t.Fatalf("move %q (%s %s) rejected: %v", m.Label, m.Type, string(m.Params), err)
	}
}

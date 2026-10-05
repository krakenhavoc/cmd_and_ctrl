package legal_test

import (
	"fmt"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// end_turn_test.go — #2165 (CR 724.1, end the turn) from the
// enumerator's side. Sundial's "activate only during your turn" is
// offered exactly when the engine accepts it; the cleanup step the
// turn skips to looks like any other — a discard owed is the only
// move, and nobody else has one; and Obeka's question goes to the
// player whose turn it is, who is the only seat offered its answers.

const (
	oracleSundial = "79d46e09-2548-440f-9c02-3c3dc39a0cd1"
	oracleObeka   = "21b76a94-d9f3-4c34-ab24-7e89321b04f0"
)

func TestSundialIsOfferedOnlyDuringYourTurnAndTheEndedTurnAgrees(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(active)
	clearHand(opp)
	mine := battlefieldCard(g, active, game.Card{Name: "Sundial of the Infinite", TypeLine: "Artifact", OracleID: oracleSundial})
	theirs := battlefieldCard(g, opp, game.Card{Name: "Sundial of the Infinite", TypeLine: "Artifact", OracleID: oracleSundial})
	mana(g, active, 1)
	mana(g, opp, 1)
	for i := 0; i < 9; i++ {
		handCard(active, creature(fmt.Sprintf("Card %d", i), "{9}", 1, 1))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, active.ID), mine)
	if len(acts) != 1 {
		t.Fatalf("want one Sundial activation on its controller's turn, got %d: %v", len(acts), labels(legal.EnumerateFor(g, active.ID)))
	}
	dispatchAll(t, g, active.ID, acts)
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(acts[0].Type), Player: active.ID, Caller: active.ID, Params: acts[0].Params}); err != nil {
		t.Fatalf("activate: %v", err)
	}

	// The opponent holds priority over the ability on someone else's
	// turn: their own Sundial is not offered, and the engine agrees.
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if !holdsPriority(g, opp.ID) {
		t.Fatalf("setup: the opponent does not hold priority")
	}
	if n := len(activationsOf(legal.EnumerateFor(g, opp.ID), theirs)); n != 0 {
		t.Errorf("Sundial offered to its controller on an opponent's turn: %d moves", n)
	}
	if err := g.ActivateCatalogAbility(opp.ID, theirs, 0, game.ActivateAbilityParams{}); err == nil {
		t.Errorf("the engine accepted an activation the enumerator does not offer")
	}

	// Everyone passes; the turn ends into a cleanup step owing a discard.
	turn := g.Turn.Seq
	for i := 0; i < 8 && g.Turn.Step != game.StepCleanup; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if g.Turn.Step != game.StepCleanup || g.Turn.Seq != turn || g.DiscardPending[active.ID] != 2 {
		t.Fatalf("at turn %d %s pending=%v, want this turn's cleanup owing two", g.Turn.Seq, g.Turn.Step, g.DiscardPending)
	}
	moves := legal.EnumerateFor(g, active.ID)
	if len(moves) == 0 {
		t.Fatal("the active player owes a discard and is offered nothing")
	}
	for _, m := range moves {
		if m.Type != legal.TypeDiscardSelection {
			t.Errorf("only discard_selection is legal at a cleanup pause: %q", m.Label)
		}
	}
	dispatchAll(t, g, active.ID, moves)
	for _, p := range g.Seats {
		if p.ID != active.ID {
			if m := legal.EnumerateFor(g, p.ID); len(m) != 0 {
				t.Errorf("%s should have no moves during the discard: %v", p.Name, labels(m))
			}
		}
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.TypeDiscardSelection, Player: active.ID, Caller: active.ID, Params: moves[0].Params}); err != nil {
		t.Fatal(err)
	}
	if g.Turn.Seq != turn+1 || g.Turn.ActiveSeat != opp.Seat {
		t.Fatalf("the discard did not end the turn: turn %d seat %d %s", g.Turn.Seq, g.Turn.ActiveSeat, g.Turn.Step)
	}

	// Now it is the opponent's turn, and their Sundial is offered.
	acts = activationsOf(legal.EnumerateFor(g, opp.ID), theirs)
	if len(acts) != 1 {
		t.Errorf("Sundial not offered on its controller's own turn: %v", labels(legal.EnumerateFor(g, opp.ID)))
	}
	dispatchAll(t, g, opp.ID, acts)
}

func TestObekaQuestionIsOfferedToThePlayerWhoseTurnItIs(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	obeka := battlefieldCard(g, opp, game.Card{
		Name: "Obeka, Brute Chronologist", TypeLine: "Legendary Creature — Ogre Wizard",
		OracleID: oracleObeka, Power: 3, Toughness: 4,
	})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	acts := activationsOf(legal.EnumerateFor(g, opp.ID), obeka)
	if len(acts) != 1 {
		t.Fatalf("want Obeka's activation offered to its controller, got %v", labels(legal.EnumerateFor(g, opp.ID)))
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(acts[0].Type), Player: opp.ID, Caller: opp.ID, Params: acts[0].Params}); err != nil {
		t.Fatalf("activate Obeka: %v", err)
	}
	for i := 0; i < 8 && len(g.PendingChoices) == 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	answers := legal.EnumerateFor(g, active.ID)
	if len(answers) != 2 {
		t.Fatalf("the active player should be offered yes and no, got %v", labels(answers))
	}
	for _, m := range answers {
		if m.Type != legal.TypeResolveChoice {
			t.Errorf("only the question's answers are legal while it is open: %q", m.Label)
		}
	}
	dispatchAll(t, g, active.ID, answers)
	if m := legal.EnumerateFor(g, opp.ID); len(m) != 0 {
		t.Errorf("Obeka's controller is offered moves while the active player decides: %v", labels(m))
	}
}

package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// passTurnBoard puts four step-trigger permanents on a two-player
// table, each of which gains its controller life:
//
//   - the active player's "at the beginning of combat on your turn,
//     gain 1 life";
//   - the active player's "at the beginning of your end step, gain 1
//     life";
//   - the active player's "at the beginning of your postcombat main
//     phase, you may gain 10 life" — the trigger that needs a decision;
//   - the opponent's "at the beginning of each end step, gain 1 life".
func passTurnBoard(t *testing.T, g *Game) {
	t.Helper()
	active, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1-g.Turn.ActiveSeat]
	const (
		combatOracle = "test-2881-combat"
		endOracle    = "test-2881-your-end"
		mayOracle    = "test-2881-may-main2"
		eachOracle   = "test-2881-each-end"
	)
	for _, c := range []struct {
		oracle, name string
		owner        *Player
	}{
		{combatOracle, "Combat Watcher", active},
		{endOracle, "End Step Watcher", active},
		{mayOracle, "Second Main Watcher", active},
		{eachOracle, "Each End Step Watcher", opp},
	} {
		g.Battlefield.PushTop(Card{
			InstanceID: uuid.New(), Name: c.name, OracleID: c.oracle,
			TypeLine: "Enchantment", Owner: c.owner.ID, Controller: c.owner.ID,
		})
	}
	gain := func(n int) func(g *Game, item *StackItem) error {
		return func(g *Game, item *StackItem) error {
			return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, n)
		}
	}
	yours := func(step Step) func(Event, *Card, Characteristic, *Game) bool {
		return func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
			return ev.Step == step && ev.Actor == source.Controller
		}
	}
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		switch id {
		case combatOracle:
			return []TriggeredAbility{{
				Watches:   []EventKind{EventStepBegan},
				AppliesTo: yours(StepBeginCombat),
				Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
					return newTriggeredItemForTest(source, "Combat Watcher — gain 1 life", gain(1))
				},
			}}
		case endOracle:
			return []TriggeredAbility{{
				Watches: []EventKind{EventBeginEndStep},
				AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
					return ev.Actor == source.Controller
				},
				Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
					return newTriggeredItemForTest(source, "End Step Watcher — gain 1 life", gain(1))
				},
			}}
		case mayOracle:
			return []TriggeredAbility{{
				Watches:        []EventKind{EventStepBegan},
				AppliesTo:      yours(StepPostcombatMain),
				OptionalPrompt: &TriggerOptionalPrompt{Question: "Gain 10 life?"},
				Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
					return newTriggeredItemForTest(source, "Second Main Watcher — gain 10 life", gain(10))
				},
			}}
		case eachOracle:
			return []TriggeredAbility{{
				Watches: []EventKind{EventBeginEndStep},
				Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
					return newTriggeredItemForTest(source, "Each End Step Watcher — gain 1 life", gain(1))
				},
			}}
		}
		return nil
	})
}

// opponentPassesUntil plays the opponent's side of the table the way
// the dispatcher does: each time they hold priority they pass, and
// SettlePassTurn runs after the action. It returns when stop reports
// true, or fails the test if the table stops moving.
func opponentPassesUntil(t *testing.T, g *Game, opp *Player, stop func() bool) {
	t.Helper()
	for i := 0; i < 200; i++ {
		if stop() {
			return
		}
		h := g.Turn.PriorityHolder
		if h < 0 || g.Seats[h].ID != opp.ID {
			t.Fatalf("the table stopped at turn %d %s with priority %d and %d prompts open",
				g.Turn.Seq, g.Turn.Step, h, len(g.PendingChoices))
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("opponent's pass at %s: %v", g.Turn.Step, err)
		}
		g.SettlePassTurn()
	}
	t.Fatalf("the table did not reach the stop in 200 passes")
}

// TestEndTurnWalksEveryStepAndFiresStepTriggers is #2881's repro:
// "pass turn" from the dock (end_turn) from the precombat main phase must still begin every step
// left in the turn (CR 500.1), put each step's triggers on the stack
// the next time a player would receive priority (CR 500.6, 603.3) and
// resolve them, and stop for a trigger that needs its controller's
// decision. The opponent keeps priority in every step (CR 117.3d) and
// their "each end step" trigger fires too.
func TestEndTurnWalksEveryStepAndFiresStepTriggers(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	active, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1-g.Turn.ActiveSeat]
	passTurnBoard(t, g)
	seq := g.Turn.Seq
	activeLife, oppLife := active.Life, opp.Life

	if err := g.EndTurnByPassing(); err != nil {
		t.Fatalf("EndTurnByPassing: %v", err)
	}
	if g.Turn.Seq != seq {
		t.Fatalf("end turn jumped to turn %d: the rest of turn %d never happened (CR 500.1)", g.Turn.Seq, seq)
	}

	// Beginning of combat: the active player's trigger resolves with
	// no decision; the passing stops at the "may" in the second main
	// phase and the active player is asked.
	opponentPassesUntil(t, g, opp, func() bool { return len(g.PendingChoices) > 0 })
	if g.Turn.Seq != seq || g.Turn.Step != StepPostcombatMain {
		t.Fatalf("stopped at turn %d %s, want turn %d postcombat_main for the \"may\"", g.Turn.Seq, g.Turn.Step, seq)
	}
	prompt := g.PendingChoices[0]
	if prompt.Chooser != active.ID {
		t.Fatalf("the \"may\" went to %s, want the active player %s", prompt.Chooser, active.ID)
	}
	if active.Life != activeLife+1 {
		t.Fatalf("active life %d, want %d: the beginning-of-combat trigger did not resolve", active.Life, activeLife+1)
	}
	// Nothing walks past the open prompt.
	g.SettlePassTurn()
	if len(g.PendingChoices) != 1 || g.Turn.Step != StepPostcombatMain {
		t.Fatalf("the passing walked past the open prompt to %s", g.Turn.Step)
	}

	if err := g.ResolveTriggerPrompt(prompt.ID, active.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	g.SettlePassTurn()
	opponentPassesUntil(t, g, opp, func() bool { return g.Turn.Seq != seq })

	want := []Step{StepBeginCombat, StepDeclareAttackers, StepDeclareBlockers, StepCombatDamage,
		StepEndCombat, StepPostcombatMain, StepEnd, StepCleanup}
	got := stepsBegunInTurn(g, seq)
	for _, s := range want {
		found := false
		for _, b := range got {
			if b == s {
				found = true
			}
		}
		if !found {
			t.Errorf("step %s never began in the passed turn (began: %v)", s, got)
		}
	}
	if active.Life != activeLife+12 {
		t.Errorf("active life %d, want %d: combat +1, the \"may\" +10, end step +1", active.Life, activeLife+12)
	}
	if opp.Life != oppLife+1 {
		t.Errorf("opponent life %d, want %d: their \"each end step\" trigger", opp.Life, oppLife+1)
	}
	if g.Seats[g.Turn.ActiveSeat].ID == active.ID {
		t.Errorf("the turn did not pass to the opponent")
	}
	if g.PassingTurn() {
		t.Errorf("the pass outlived its turn")
	}
	// The opponent's turn is theirs: the active player's pass is gone,
	// so they are asked again when they get priority.
	opponentPassesUntil(t, g, opp, func() bool {
		h := g.Turn.PriorityHolder
		return h >= 0 && g.Seats[h].ID == active.ID
	})
}

// TestEndTurnRefusedWhilePromptOpen: like pass_turn, a prompt the
// table waits for is answered before the turn is ended (#730).
func TestEndTurnRefusedWhilePromptOpen(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	active := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(PendingChoice{
			Kind: PendingChoiceTriggerPrompt, Chooser: active.ID, Count: 1, Reason: "test",
		})
	})
	if err := g.EndTurnByPassing(); !errors.Is(err, ErrChoicePending) {
		t.Fatalf("EndTurnByPassing with a prompt open = %v, want ErrChoicePending", err)
	}
	if g.PassingTurn() {
		t.Fatalf("a refused end turn left a standing pass")
	}
}

// TestUndoTakesEndTurnBack: the standing pass is part of the game a
// clone carries, so restoring the pre-pass clone ends it.
func TestUndoTakesEndTurnBack(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	pre := g.Clone()
	if err := g.EndTurnByPassing(); err != nil {
		t.Fatalf("EndTurnByPassing: %v", err)
	}
	if !g.PassingTurn() {
		t.Fatalf("setup: no standing pass after EndTurnByPassing")
	}
	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	if g.PassingTurn() {
		t.Fatalf("the undone end turn is still standing")
	}
	if err := g.EndTurnByPassing(); err != nil {
		t.Fatalf("EndTurnByPassing: %v", err)
	}
	if c := g.Clone(); !c.PassingTurn() {
		t.Fatalf("a clone dropped the standing pass")
	}
}

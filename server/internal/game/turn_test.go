package game

import "testing"

func TestTurnSequenceLength(t *testing.T) {
	seq := TurnSequence()
	// Thirteen since #717: the first-strike combat damage step
	// (CR 510.4) is in the sequence like any other, and is skipped in
	// the turns that do not have it.
	if len(seq) != 13 {
		t.Errorf("turn sequence: got %d steps, want 13", len(seq))
	}
}

func TestTurnSequenceExactOrder(t *testing.T) {
	want := []Step{
		StepUntap,
		StepUpkeep,
		StepDraw,
		StepPrecombatMain,
		StepBeginCombat,
		StepDeclareAttackers,
		StepDeclareBlockers,
		StepFirstStrikeDamage,
		StepCombatDamage,
		StepEndCombat,
		StepPostcombatMain,
		StepEnd,
		StepCleanup,
	}
	got := TurnSequence()
	if len(got) != len(want) {
		t.Fatalf("length: got %d, want %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("step %d: got %q, want %q", i, got[i], w)
		}
	}
}

func TestTurnSequenceReturnsCopy(t *testing.T) {
	seq := TurnSequence()
	seq[0] = "clobbered"
	again := TurnSequence()
	if again[0] == "clobbered" {
		t.Error("TurnSequence leaked its internal slice")
	}
}

func TestPhaseOfMapping(t *testing.T) {
	cases := []struct {
		step  Step
		phase Phase
	}{
		{StepUntap, PhaseBeginning},
		{StepUpkeep, PhaseBeginning},
		{StepDraw, PhaseBeginning},
		{StepPrecombatMain, PhasePrecombatMain},
		{StepBeginCombat, PhaseCombat},
		{StepDeclareAttackers, PhaseCombat},
		{StepDeclareBlockers, PhaseCombat},
		{StepFirstStrikeDamage, PhaseCombat},
		{StepCombatDamage, PhaseCombat},
		{StepEndCombat, PhaseCombat},
		{StepPostcombatMain, PhasePostcombatMain},
		{StepEnd, PhaseEnding},
		{StepCleanup, PhaseEnding},
	}
	for _, c := range cases {
		if got := PhaseOf(c.step); got != c.phase {
			t.Errorf("PhaseOf(%q): got %q, want %q", c.step, got, c.phase)
		}
	}
}

func TestTurnAdvanceWithinTurn(t *testing.T) {
	turn := Turn{Number: 1, ActiveSeat: 0, Phase: PhaseBeginning, Step: StepUntap}
	turn = turn.advance(4)
	if turn.Step != StepUpkeep {
		t.Errorf("after untap: got %q, want %q", turn.Step, StepUpkeep)
	}
	if turn.Number != 1 || turn.ActiveSeat != 0 {
		t.Errorf("turn/seat changed mid-turn: number=%d seat=%d", turn.Number, turn.ActiveSeat)
	}
}

func TestTurnAdvanceWrapsToNextSeat(t *testing.T) {
	// Start at the last step of seat 0 on turn 1.
	turn := Turn{Number: 1, ActiveSeat: 0, Phase: PhaseEnding, Step: StepCleanup}
	turn = turn.advance(4)
	if turn.Step != StepUntap {
		t.Errorf("after cleanup: got %q, want %q", turn.Step, StepUntap)
	}
	if turn.ActiveSeat != 1 {
		t.Errorf("seat after wrap: got %d, want 1", turn.ActiveSeat)
	}
	if turn.Number != 1 {
		t.Errorf("turn number: got %d, want 1 (still on turn 1 until all 4 seats play)", turn.Number)
	}
}

func TestTurnAdvanceWrapsToNextRound(t *testing.T) {
	// Last step of the last seat on turn 1 → seat 0, turn 2.
	turn := Turn{Number: 1, ActiveSeat: 3, Phase: PhaseEnding, Step: StepCleanup}
	turn = turn.advance(4)
	if turn.ActiveSeat != 0 {
		t.Errorf("seat after full round: got %d, want 0", turn.ActiveSeat)
	}
	if turn.Number != 2 {
		t.Errorf("turn number after full round: got %d, want 2", turn.Number)
	}
	if turn.Step != StepUntap {
		t.Errorf("step after full round: got %q, want %q", turn.Step, StepUntap)
	}
}

func TestIndexOfStepUnknownReturnsNegative(t *testing.T) {
	if idx := indexOfStep("nonsense"); idx != -1 {
		t.Errorf("indexOfStep(nonsense): got %d, want -1", idx)
	}
}

func TestTurnAdvanceIntoUntapSetsNoPriority(t *testing.T) {
	// Cleanup → next seat's Untap must land with PriorityHolder set to
	// the NoPriority sentinel — S13 no-priority-on-untap guarantee.
	turn := Turn{Number: 1, ActiveSeat: 0, Phase: PhaseEnding, Step: StepCleanup}
	turn = turn.advance(4)
	if turn.Step != StepUntap {
		t.Fatalf("after cleanup: got %q, want %q", turn.Step, StepUntap)
	}
	if turn.PriorityHolder != NoPriority {
		t.Errorf("PriorityHolder on Untap: got %d, want NoPriority (%d)", turn.PriorityHolder, NoPriority)
	}
}

func TestTurnAdvanceIntoCleanupSetsNoPriority(t *testing.T) {
	turn := Turn{Number: 1, ActiveSeat: 0, Phase: PhaseEnding, Step: StepEnd, PriorityHolder: 0}
	turn = turn.advance(4)
	if turn.Step != StepCleanup {
		t.Fatalf("after end: got %q, want %q", turn.Step, StepCleanup)
	}
	if turn.PriorityHolder != NoPriority {
		t.Errorf("PriorityHolder on Cleanup: got %d, want NoPriority (%d)", turn.PriorityHolder, NoPriority)
	}
}

func TestTurnAdvanceIntoPriorityStepsSetsActiveSeat(t *testing.T) {
	// Every other step must grant priority to the active seat.
	cases := []struct {
		from Step
		want Step
	}{
		{StepUntap, StepUpkeep},
		{StepUpkeep, StepDraw},
		{StepDraw, StepPrecombatMain},
		{StepPrecombatMain, StepBeginCombat},
		{StepBeginCombat, StepDeclareAttackers},
		{StepDeclareAttackers, StepDeclareBlockers},
		{StepDeclareBlockers, StepFirstStrikeDamage},
		{StepFirstStrikeDamage, StepCombatDamage},
		{StepCombatDamage, StepEndCombat},
		{StepEndCombat, StepPostcombatMain},
		{StepPostcombatMain, StepEnd},
	}
	for _, c := range cases {
		turn := Turn{Number: 1, ActiveSeat: 2, Phase: PhaseOf(c.from), Step: c.from}
		got := turn.advance(4)
		if got.Step != c.want {
			t.Errorf("advance from %q: got step %q, want %q", c.from, got.Step, c.want)
		}
		if got.PriorityHolder != 2 {
			t.Errorf("advance from %q: PriorityHolder=%d, want ActiveSeat=2", c.from, got.PriorityHolder)
		}
	}
}

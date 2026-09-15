package game

import "testing"

// step_event_test.go — #588: EventStepBegan carries the typed step and
// is emitted exactly once per step that begins, including turn 1's
// untap, which the mulligan window used to announce once per keep.

func TestStepBeganCarriesTheStepAndFiresOncePerStep(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	untaps := 0
	for _, ev := range g.Events {
		if ev.Kind == EventStepBegan && ev.Step == StepUntap && ev.Amount == 1 {
			untaps++
		}
		if ev.Kind == EventStepBegan && ev.Step == "" {
			t.Fatalf("EventStepBegan without a typed Step: %+v", ev)
		}
		if ev.Kind == EventStepBegan && string(ev.Step) != ev.Label {
			t.Fatalf("Step %q and Label %q disagree", ev.Step, ev.Label)
		}
	}
	if untaps != 1 {
		t.Errorf("turn 1 untap announced %d times, want 1", untaps)
	}
	seen := 0
	for g.Turn.Step != StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for _, ev := range g.Events {
		if ev.Kind == EventStepBegan && ev.Step == StepBeginCombat {
			seen++
		}
	}
	if seen != 1 {
		t.Errorf("begin combat announced %d times, want 1", seen)
	}
}

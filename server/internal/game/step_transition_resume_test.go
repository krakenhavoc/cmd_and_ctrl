package game

import "testing"

// step_transition_resume_test.go is the #710 regression suite.
//
// A step transition runs through the CR 614 replacement pipeline like
// every other event, and like every other event it can PAUSE: two
// applicable replacements make CR 616 ask the affected player to order
// them. The resume for that pause was `return nil`, and
// runStepEntryHooksLocked fell through to running the step, so two
// "skip your draw step" effects under one controller queued a prompt
// and then drew a card anyway — the opposite of what both cards say.
//
// Two halves are tested here:
//
//   - the SHORT-CIRCUIT: two effects that do nothing but cancel have
//     no ordering worth asking about (CR 616.1 is unobservable), so no
//     prompt is queued at all;
//   - the RESUME: when something else is in the mix the prompt is
//     real, and the step must not run until it is answered — then it
//     is skipped or performed exactly as the unpaused path would.

// stepPos is one position of the turn cursor.
type stepPos struct {
	Seat int
	Step Step
}

// skipStepForSeat builds a replacement that cancels entry into `step`
// for one seat — the shape of "skip your draw step". `pure` sets the
// PureCancel declaration; a test that wants the CR 616 prompt to
// actually queue leaves it off.
func skipStepForSeat(step Step, seat int, pure bool, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventStepTransition},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventStepTransition &&
				ev.StepTransitionStep == step && ev.StepTransitionSeat == seat
		},
		Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
			ev.Cancel()
			return nil
		},
		PureCancel: pure,
		Label:      label,
	}
}

// watchStepForSeat builds a replacement that applies to the same
// transition and changes nothing. Not a cancel, so a pair containing
// it is a real CR 616 ordering decision — the mixed case #710 says
// must never silently run the step.
func watchStepForSeat(step Step, seat int, label string) ReplacementEffect {
	return ReplacementEffect{
		Watches: []EventKind{EventStepTransition},
		AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
			return ev.Kind == RepEventStepTransition &&
				ev.StepTransitionStep == step && ev.StepTransitionSeat == seat
		},
		Replace: func(*ReplacementEvent, *Game, *Card) error { return nil },
		Label:   label,
	}
}

// newGameWithStepReplacements returns a started two-seat game with the
// given step-transition replacements registered.
func newGameWithStepReplacements(t *testing.T, effects ...ReplacementEffect) *Game {
	t.Helper()
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		for _, e := range effects {
			g.RegisterReplacementForTest(e)
		}
	})
	return g
}

// advanceUntil walks the turn cursor, recording every position it
// lands on, and stops when done() is satisfied OR a prompt appears.
// Stopping on a prompt is the point: a step-transition pause parks the
// cursor, and advancing past it would strand the prompt.
//
// A trace rather than a spot check for the reason the catalog tests
// give: "the draw step was skipped" is a statement about a step that
// never happened, and there is no moment at which the cursor shows it.
func advanceUntil(t *testing.T, g *Game, limit int, done func() bool) []stepPos {
	t.Helper()
	trace := []stepPos{{Seat: g.Turn.ActiveSeat, Step: g.Turn.Step}}
	for i := 0; i < limit; i++ {
		if done() || len(g.PendingChoices) > 0 {
			return trace
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep iter %d: %v", i, err)
		}
		trace = append(trace, stepPos{Seat: g.Turn.ActiveSeat, Step: g.Turn.Step})
	}
	t.Fatalf("turn cursor never reached the wanted position in %d steps (trace %v)", limit, trace)
	return nil
}

// sawPos reports whether the trace ever landed on a seat's step.
func sawPos(trace []stepPos, seat int, step Step) bool {
	for _, f := range trace {
		if f.Seat == seat && f.Step == step {
			return true
		}
	}
	return false
}

// walkToSeat1PrecombatMain drives the cursor from where
// newActiveGame leaves it (seat 0's upkeep) to seat 1's precombat
// main, which is one step past the draw step under test. Seat 1
// rather than seat 0 because CR 103.8a already skips the starting
// seat's turn-1 draw step in a two-player game, so a skip asserted
// there would prove nothing.
func walkToSeat1PrecombatMain(t *testing.T, g *Game) []stepPos {
	t.Helper()
	return advanceUntil(t, g, 40, func() bool {
		return g.Turn.ActiveSeat == 1 && g.Turn.Step == StepPrecombatMain
	})
}

// TestOneSkipStepReplacementSkipsTheStep is the control: nothing about
// the single-applicable fast path changes, prompt or no prompt.
func TestOneSkipStepReplacementSkipsTheStep(t *testing.T) {
	g := newGameWithStepReplacements(t, skipStepForSeat(StepDraw, 1, true, "skip draw"))
	hand := g.Seats[1].Hand.Size()

	trace := walkToSeat1PrecombatMain(t, g)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("one replacement queued %d prompts, want 0", len(g.PendingChoices))
	}
	if sawPos(trace, 1, StepDraw) {
		t.Error("seat 1's draw step happened despite the skip")
	}
	if got := g.Seats[1].Hand.Size(); got != hand {
		t.Errorf("seat 1 drew %d cards through a skipped draw step", got-hand)
	}
}

// TestTwoPureCancelStepSkipsNeedNoPromptAndStillSkip is the headline
// case: Necropotence and Yawgmoth's Bargain under one controller.
// Both cancel and nothing else, so CR 616 has no ordering to ask
// about — the engine applies them without prompting, and the step is
// skipped once.
func TestTwoPureCancelStepSkipsNeedNoPromptAndStillSkip(t *testing.T) {
	g := newGameWithStepReplacements(t,
		skipStepForSeat(StepDraw, 1, true, "skip draw A"),
		skipStepForSeat(StepDraw, 1, true, "skip draw B"),
	)
	hand := g.Seats[1].Hand.Size()

	trace := walkToSeat1PrecombatMain(t, g)

	if len(g.PendingChoices) != 0 {
		t.Fatalf("two pure cancels queued %d prompts, want 0 — every ordering "+
			"skips the same step, so there is nothing to order (CR 616.1)",
			len(g.PendingChoices))
	}
	if sawPos(trace, 1, StepDraw) {
		t.Error("seat 1's draw step happened despite two skip-step replacements")
	}
	if got := g.Seats[1].Hand.Size(); got != hand {
		t.Errorf("seat 1 drew %d cards through a doubly-skipped draw step (#710)", got-hand)
	}
}

// TestPausedStepTransitionDoesNotRunTheStepBeforeTheAnswer — a cancel
// paired with something that is not a cancel is a real CR 616 prompt.
// The step must not run while it is outstanding: before #710 the pause
// site fell through and the draw happened with the prompt still on
// screen, so answering it changed nothing.
//
// Both answer orders are exercised, because both have to end in a
// skipped step: cancel-first ends the apply-loop immediately, and
// cancel-second still cancels before the pipeline settles.
func TestPausedStepTransitionDoesNotRunTheStepBeforeTheAnswer(t *testing.T) {
	for _, tc := range []struct {
		name    string
		reverse bool
	}{
		{name: "answered in the offered order"},
		{name: "answered cancel-first", reverse: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newGameWithStepReplacements(t,
				watchStepForSeat(StepDraw, 1, "watch draw"),
				skipStepForSeat(StepDraw, 1, false, "skip draw"),
			)
			hand := g.Seats[1].Hand.Size()

			trace := walkToSeat1PrecombatMain(t, g)

			if len(g.PendingChoices) != 1 {
				t.Fatalf("pending choices = %d, want 1 CR 616 ordering prompt (trace %v)",
					len(g.PendingChoices), trace)
			}
			prompt := g.PendingChoices[0]
			if prompt.Kind != PendingChoiceReplacementOrder {
				t.Fatalf("prompt kind = %q, want %q", prompt.Kind, PendingChoiceReplacementOrder)
			}
			if prompt.Chooser != g.Seats[1].ID {
				t.Errorf("chooser = %s, want seat 1 (%s) — the affected player", prompt.Chooser, g.Seats[1].ID)
			}
			if g.Turn.ActiveSeat != 1 || g.Turn.Step != StepDraw {
				t.Fatalf("cursor at seat %d %q while the prompt is pending, want seat 1 %q",
					g.Turn.ActiveSeat, g.Turn.Step, StepDraw)
			}
			if got := g.Seats[1].Hand.Size(); got != hand {
				t.Fatalf("seat 1 drew %d cards while the ordering prompt was still "+
					"unanswered — the pause must not fall through to the step (#710)", got-hand)
			}

			order := append([]ReplacementEffectID(nil), prompt.ReplacementEffectIDs...)
			if tc.reverse {
				order[0], order[1] = order[1], order[0]
			}
			if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, order); err != nil {
				t.Fatalf("ResolveReplacementOrder: %v", err)
			}

			if len(g.PendingChoices) != 0 {
				t.Errorf("%d choices still queued after the answer, want 0", len(g.PendingChoices))
			}
			if got := g.Seats[1].Hand.Size(); got != hand {
				t.Errorf("seat 1 drew %d cards after answering — the step is cancelled, so "+
					"the resume skips it (CR 500.11)", got-hand)
			}
			if g.Turn.Step != StepPrecombatMain || g.Turn.ActiveSeat != 1 {
				t.Errorf("cursor at seat %d %q after the answer, want seat 1 %q — a cancelled "+
					"step transition still has to move the cursor past the skipped step",
					g.Turn.ActiveSeat, g.Turn.Step, StepPrecombatMain)
			}
		})
	}
}

// TestPausedStepTransitionRunsTheStepWhenNothingCancelled — the other
// half of the resume. Two replacements that both leave the transition
// alone still prompt, and the answer has to PERFORM the step's
// turn-based action: exactly one draw, once, after the answer. A
// resume that only handled the cancel branch would eat the draw step
// of anyone with two harmless step replacements in play.
func TestPausedStepTransitionRunsTheStepWhenNothingCancelled(t *testing.T) {
	g := newGameWithStepReplacements(t,
		watchStepForSeat(StepDraw, 1, "watch draw A"),
		watchStepForSeat(StepDraw, 1, "watch draw B"),
	)
	hand := g.Seats[1].Hand.Size()

	walkToSeat1PrecombatMain(t, g)

	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1", len(g.PendingChoices))
	}
	prompt := g.PendingChoices[0]
	if got := g.Seats[1].Hand.Size(); got != hand {
		t.Fatalf("seat 1 drew %d cards before answering the prompt", got-hand)
	}
	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder: %v", err)
	}
	if got := g.Seats[1].Hand.Size(); got != hand+1 {
		t.Errorf("seat 1's hand = %d, want %d — nothing cancelled the transition, so the "+
			"resume owes the draw step its turn-based action", got, hand+1)
	}
	if g.Turn.ActiveSeat != 1 || g.Turn.Step != StepDraw {
		t.Errorf("cursor at seat %d %q, want seat 1 %q — the draw step grants priority, "+
			"so it does not auto-advance", g.Turn.ActiveSeat, g.Turn.Step, StepDraw)
	}
}

// TestStaleStepTransitionResumeIsDropped — the prompt parks the
// cursor, but nothing in the engine refuses an advance_step while it
// is outstanding. If the table walks on and the prompt is answered
// afterwards, finishing it would skip or run a step nobody is in.
// The answer is dropped instead, leaving the cursor where the players
// put it.
func TestStaleStepTransitionResumeIsDropped(t *testing.T) {
	g := newGameWithStepReplacements(t,
		watchStepForSeat(StepDraw, 1, "watch draw"),
		skipStepForSeat(StepDraw, 1, false, "skip draw"),
	)
	walkToSeat1PrecombatMain(t, g)
	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1", len(g.PendingChoices))
	}
	prompt := g.PendingChoices[0]

	// The table advances past the paused step without answering.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	moved := stepPos{Seat: g.Turn.ActiveSeat, Step: g.Turn.Step}
	if moved.Step != StepPrecombatMain {
		t.Fatalf("cursor at %q after advancing past the prompt, want %q", moved.Step, StepPrecombatMain)
	}

	if err := g.ResolveReplacementOrder(prompt.ID, prompt.Chooser, prompt.ReplacementEffectIDs); err != nil {
		t.Fatalf("ResolveReplacementOrder on a stale prompt = %v, want nil — the prompt is "+
			"already dequeued, so it must not fail the action either", err)
	}
	if got := (stepPos{Seat: g.Turn.ActiveSeat, Step: g.Turn.Step}); got != moved {
		t.Errorf("cursor moved to seat %d %q answering a stale step prompt, want it left at seat %d %q",
			got.Seat, got.Step, moved.Seat, moved.Step)
	}
}

// TestPureCancelIsIgnoredWhenTheEffectAlsoAsksAQuestion — the
// short-circuit fires Replace without prompting, which would swallow a
// CR 614.10 "may" decision. An Optional effect is never treated as a
// pure cancel however it is declared, so the ordering prompt still
// queues and the "may" prompt still gets asked.
func TestPureCancelIsIgnoredWhenTheEffectAlsoAsksAQuestion(t *testing.T) {
	optional := skipStepForSeat(StepDraw, 1, true, "may skip draw")
	optional.Optional = true
	g := newGameWithStepReplacements(t, optional, skipStepForSeat(StepDraw, 1, true, "skip draw"))

	walkToSeat1PrecombatMain(t, g)

	if len(g.PendingChoices) != 1 {
		t.Fatalf("pending choices = %d, want 1 — an Optional replacement is not a pure "+
			"cancel, so the ordering prompt stands", len(g.PendingChoices))
	}
	if got := g.PendingChoices[0].Kind; got != PendingChoiceReplacementOrder {
		t.Errorf("prompt kind = %q, want %q", got, PendingChoiceReplacementOrder)
	}
}

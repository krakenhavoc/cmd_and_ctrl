package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// loop_shortcut_test.go — #804, CR 726, the half ADR 0055 §6 deferred.
//
// The breaker stops a loop and hands priority back. On its own that
// leaves the table clicking "next" once per iteration, or turning
// autopass back on to be stopped again a threshold later. CR 726 is
// the conversation paper has instead: the loop's controller names a
// number of further iterations and the table skips there.
//
// These drive the same real two-permanent loop the breaker tests do
// (seedTriggerLoop), through the real priority loop, and assert the
// number means exactly what it says.

// resolutionsOf counts how many times the Spark engine's ability has
// resolved this turn — the raw per-turn tally, which no decision
// resets, so it measures iterations rather than runs.
func resolutionsOf(g *Game, source uuid.UUID, label string) int {
	return g.TurnTally.Resolved[TallyKey(source, label)]
}

// TestShortcutPromptAppearsWithTheNotice — the prompt is queued to the
// loop's CONTROLLER, at the same moment the notice goes up, carrying
// the count the notice names and the key the answer attaches to.
func TestShortcutPromptAppearsWithTheNotice(t *testing.T) {
	const threshold = 4
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	_, spark := seedTriggerLoop(t, g)

	// Nothing before the threshold.
	for i := 1; i < threshold; i++ {
		passAroundOnce(t, g)
		passAroundOnce(t, g)
		if c := loopShortcutPrompt(g); c != nil {
			t.Fatalf("shortcut prompt after %d resolutions, want none below %d: %+v", i, threshold, c)
		}
	}

	passAroundOnce(t, g)
	c := loopShortcutPrompt(g)
	if c == nil {
		t.Fatalf("no shortcut prompt with the notice up after %d resolutions", threshold)
	}
	if c.Chooser != g.Seats[0].ID {
		t.Errorf("prompt chooser = %s, want the loop's controller %s", c.Chooser, g.Seats[0].ID)
	}
	if c.Source != spark {
		t.Errorf("prompt source = %s, want the Spark engine %s", c.Source, spark)
	}
	if c.Reason != loopLabelSpark {
		t.Errorf("prompt reason = %q, want the ability's label %q", c.Reason, loopLabelSpark)
	}
	if c.LoopShortcutCount != threshold {
		t.Errorf("prompt count = %d, want %d — the N in \"has resolved N times\"", c.LoopShortcutCount, threshold)
	}
	if c.LoopShortcutKey != TallyKey(spark, loopLabelSpark) {
		t.Errorf("prompt key = %q, want the tally key %q", c.LoopShortcutKey, TallyKey(spark, loopLabelSpark))
	}
	if c.LoopShortcutRepeat {
		t.Error("the first ask of a turn is marked a repeat")
	}
	// Exactly one conversation: stepping the loop on by hand does not
	// queue a second prompt.
	if n := len(g.PendingChoices); n != 1 {
		t.Fatalf("pending choices = %d, want just the shortcut prompt", n)
	}
}

// TestShortcutRunsExactlyKMoreThenAsksAgain is the headline. K = 3
// means three more resolutions of that ability with no notice and no
// hold, and then the breaker is back with the same question.
func TestShortcutRunsExactlyKMoreThenAsksAgain(t *testing.T) {
	const (
		threshold = 4
		k         = 3
	)
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	_, spark := seedTriggerLoop(t, g)
	passUntilLoopNotice(t, g)

	at := resolutionsOf(g, spark, loopLabelSpark)
	if at != threshold {
		t.Fatalf("the notice went up after %d resolutions, want %d", at, threshold)
	}
	answerLoopShortcut(t, g, k)

	// The answer lifts the hold: automatic passers run again.
	if g.AutoPassSuspended() {
		t.Fatal("automatic passing is still suspended after the controller agreed to a shortcut")
	}
	if got := g.LoopAllowanceFor(TallyKey(spark, loopLabelSpark)); got != k {
		t.Fatalf("allowance = %d, want %d", got, k)
	}

	// Three more resolutions of the tripping ability, silently.
	for i := 1; i <= k; i++ {
		passAroundOnce(t, g) // spark resolves
		if i < k {
			if g.LoopNotice != nil {
				t.Fatalf("notice re-raised after %d of %d shortcut iterations: %+v", i, k, g.LoopNotice)
			}
			if c := loopShortcutPrompt(g); c != nil {
				t.Fatalf("re-prompted after %d of %d shortcut iterations: %+v", i, k, c)
			}
		}
		passAroundOnce(t, g) // the loop's other half resolves
	}

	if got := resolutionsOf(g, spark, loopLabelSpark); got != at+k {
		t.Errorf("the ability resolved %d times in total, want %d (%d + K=%d)", got, at+k, at, k)
	}
	if g.LoopNotice == nil {
		t.Fatal("the notice did not come back when the shortcut ran out")
	}
	if !g.AutoPassSuspended() {
		t.Error("automatic passing is live with the shortcut spent")
	}
	c := loopShortcutPrompt(g)
	if c == nil {
		t.Fatal("the question was not asked again when the shortcut ran out")
	}
	if !c.LoopShortcutRepeat {
		t.Error("the second ask of the turn is not marked a repeat — a bot would shortcut the same loop forever")
	}
	if got := g.LoopAllowanceFor(TallyKey(spark, loopLabelSpark)); got != 0 {
		t.Errorf("allowance = %d after the shortcut ran out, want 0", got)
	}
	// One breadcrumb per raise: the first and this one.
	if n := countLoopEvents(g); n != 2 {
		t.Errorf("EventLoopSuspected count = %d, want 2 (one per raise)", n)
	}
}

// TestShortcutStopNowLeavesTheTablePaused — K = 0 is "stop here". The
// prompt clears, the table unblocks, and everything else stays exactly
// as the breaker left it: notice up, automatic passing suspended.
func TestShortcutStopNowLeavesTheTablePaused(t *testing.T) {
	const threshold = 3
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	_, spark := seedTriggerLoop(t, g)
	passUntilLoopNotice(t, g)
	want := *g.LoopNotice

	answerLoopShortcut(t, g, 0)

	if loopShortcutPrompt(g) != nil {
		t.Error("the prompt survived its answer")
	}
	if g.LoopNotice == nil || *g.LoopNotice != want {
		t.Fatalf("notice = %+v after stop-here, want it unchanged at %+v", g.LoopNotice, want)
	}
	if !g.AutoPassSuspended() {
		t.Error("automatic passing resumed after stop-here")
	}
	if got := g.LoopAllowanceFor(TallyKey(spark, loopLabelSpark)); got != 0 {
		t.Errorf("allowance = %d after stop-here, want none", got)
	}
	// The table can still be stepped on by hand, which is the whole
	// of ADR 0055 §4, and the banner's count keeps up with it.
	before := g.LoopNotice.Count
	passAroundOnce(t, g)
	passAroundOnce(t, g)
	if g.LoopNotice == nil {
		t.Fatal("a bare pass cleared the notice")
	}
	if g.LoopNotice.Count <= before {
		t.Errorf("notice count = %d after another resolution, want > %d", g.LoopNotice.Count, before)
	}
	// And it does not ask again while the notice it already asked
	// about is still standing.
	if c := loopShortcutPrompt(g); c != nil {
		t.Errorf("re-prompted while the table was stepping the loop by hand: %+v", c)
	}
}

// TestShortcutIsEndedByARealDecision — mid-allowance, a cast (or any
// other decision) clears everything, exactly as it does without a
// shortcut. The number was named against a board that has now changed.
func TestShortcutIsEndedByARealDecision(t *testing.T) {
	const threshold = 3
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	_, spark := seedTriggerLoop(t, g)
	key := TallyKey(spark, loopLabelSpark)
	passUntilLoopNotice(t, g)
	answerLoopShortcut(t, g, 50)
	if got := g.LoopAllowanceFor(key); got != 50 {
		t.Fatalf("allowance = %d, want 50", got)
	}

	p := g.Seats[0]
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventCast, Actor: p.ID, CardID: uuid.New()})
	})

	if got := g.LoopAllowanceFor(key); got != 0 {
		t.Errorf("allowance = %d after a cast, want it cleared with everything else", got)
	}
	if len(g.TurnTally.LoopRun) != 0 {
		t.Errorf("loop run = %v after a cast, want empty", g.TurnTally.LoopRun)
	}
	if g.LoopNotice != nil {
		t.Errorf("notice survived a cast: %+v", g.LoopNotice)
	}
	// And the loop has to earn a fresh notice from zero.
	for i := 0; i < 2*(threshold-1); i++ {
		passAroundOnce(t, g)
	}
	if g.LoopNotice != nil {
		t.Errorf("notice raised before the run reached the threshold again: %+v", g.LoopNotice)
	}
}

// TestShortcutAnswerSurvivesUndo — the prompt and the allowance are
// per-turn state an undo must be able to rewind past. Rewinding to
// before the answer puts the question back.
func TestShortcutAnswerSurvivesUndo(t *testing.T) {
	const threshold = 3
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	_, spark := seedTriggerLoop(t, g)
	key := TallyKey(spark, loopLabelSpark)
	passUntilLoopNotice(t, g)

	asked := g.Clone()
	answerLoopShortcut(t, g, 5)
	if loopShortcutPrompt(g) != nil {
		t.Fatal("setup: the prompt survived its answer")
	}
	if got := g.LoopAllowanceFor(key); got != 5 {
		t.Fatalf("setup: allowance = %d, want 5", got)
	}

	g.WithWriteLock(func() { g.RestoreFrom(asked) })

	c := loopShortcutPrompt(g)
	if c == nil {
		t.Fatal("undo did not put the question back")
	}
	if c.LoopShortcutKey != key {
		t.Errorf("restored prompt key = %q, want %q", c.LoopShortcutKey, key)
	}
	if got := g.LoopAllowanceFor(key); got != 0 {
		t.Errorf("allowance = %d after undoing the answer, want none", got)
	}
	if g.LoopNotice == nil {
		t.Error("the notice did not come back with the prompt")
	}
	// The restored question is answerable — the ID is the one the
	// restored queue carries, and answering it grants the allowance
	// against the restored run.
	if err := g.ResolveLoopShortcut(c.ID, c.Chooser, 2); err != nil {
		t.Fatalf("ResolveLoopShortcut on the restored prompt: %v", err)
	}
	if got := g.LoopAllowanceFor(key); got != 2 {
		t.Errorf("allowance = %d after re-answering, want 2", got)
	}
}

// TestShortcutPromptSurvivesCloneAndSnapshot — the prompt's three
// fields and the allowance ride both copies, like every other piece of
// the breaker's state (snapshot_drift_test.go classifies them
// `carried`).
func TestShortcutPromptSurvivesCloneAndSnapshot(t *testing.T) {
	const threshold = 3
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	_, spark := seedTriggerLoop(t, g)
	key := TallyKey(spark, loopLabelSpark)
	passUntilLoopNotice(t, g)
	want := *loopShortcutPrompt(g)

	clone := g.Clone()
	got := loopShortcutPrompt(clone)
	if got == nil {
		t.Fatal("the clone lost the prompt")
	}
	if got.LoopShortcutKey != want.LoopShortcutKey ||
		got.LoopShortcutCount != want.LoopShortcutCount ||
		got.LoopShortcutRepeat != want.LoopShortcutRepeat {
		t.Errorf("clone prompt = %+v, want the CR 726 fields of %+v", got, want)
	}

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	got = loopShortcutPrompt(restored)
	if got == nil {
		t.Fatal("the snapshot lost the prompt")
	}
	if got.LoopShortcutKey != want.LoopShortcutKey ||
		got.LoopShortcutCount != want.LoopShortcutCount ||
		got.LoopShortcutRepeat != want.LoopShortcutRepeat {
		t.Errorf("restored prompt = %+v, want the CR 726 fields of %+v", got, want)
	}

	// The allowance too: a restore that dropped it would hand the
	// table back with a shortcut nobody is counting down.
	answerLoopShortcut(t, g, 7)
	if n := g.Clone().TurnTally.LoopAllowance[key]; n != 7 {
		t.Errorf("clone allowance = %d, want 7", n)
	}
	restored, err = g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore after the answer: %v", err)
	}
	if n := restored.TurnTally.LoopAllowance[key]; n != 7 {
		t.Errorf("restored allowance = %d, want 7", n)
	}
}

// TestShortcutRefusesAnAnswerOutOfRange — the number comes off a
// client field, so the ceiling is enforced where every other prompt's
// payload is.
func TestShortcutRefusesAnAnswerOutOfRange(t *testing.T) {
	const threshold = 3
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	seedTriggerLoop(t, g)
	passUntilLoopNotice(t, g)
	c := loopShortcutPrompt(g)

	for _, k := range []int{-1, MaxLoopShortcutIterations + 1} {
		if err := g.ResolveLoopShortcut(c.ID, c.Chooser, k); !errors.Is(err, ErrInvalidParam) {
			t.Errorf("ResolveLoopShortcut(%d) = %v, want ErrInvalidParam", k, err)
		}
	}
	if loopShortcutPrompt(g) == nil {
		t.Error("a refused answer took the prompt away")
	}
	// And it is the chooser's question, nobody else's.
	if err := g.ResolveLoopShortcut(c.ID, g.Seats[1].ID, 1); !errors.Is(err, ErrNotTheChooser) {
		t.Errorf("another seat's answer = %v, want ErrNotTheChooser", err)
	}
}

// TestShortcutPromptIsNotQueuedWithoutAControllerToAsk — the prompt
// blocks the table, so one addressed to nobody is a wedge. The notice
// still goes up; it is just not a conversation.
func TestShortcutPromptIsNotQueuedWithoutAControllerToAsk(t *testing.T) {
	g := newActiveGame(t)
	g.LoopThreshold = 2
	const label = "Ownerless Engine — tick"
	source := uuid.New()
	g.WithWriteLock(func() {
		for i := 0; i < 2; i++ {
			// Actor uuid.Nil: no seat controls this, so there is
			// nobody CR 726 would have name a number.
			g.EmitEvent(Event{Kind: EventResolve, Source: source, Label: label})
		}
	})
	if g.LoopNotice == nil {
		t.Fatal("no notice for a loop with no controller")
	}
	if c := loopShortcutPrompt(g); c != nil {
		t.Fatalf("queued a blocking prompt nobody can answer: %+v", c)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Errorf("AdvanceStep = %v, want the table free to move", err)
	}
}

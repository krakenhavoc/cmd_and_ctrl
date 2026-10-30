package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// loop_recording_test.go — #4736, ADR 0193's 2026-10-30 amendment: a
// player can ask to record a loop at any point of a game they are in,
// partway through resolving an ability included. The recording itself
// runs from one quiet priority of the recorder's to the next (CR
// 732.2a ends a shortcut where a player has priority, and the closure
// check compares two such points), so a request made where they cannot
// start is armed and begins there. Proposing still needs priority.
//
// Each test drives a real resolution: an ability on the stack that, as
// it resolves, asks its controller "gain 1 life?" (resolveUntilPrompt
// leaves it paused on that prompt). The tests call the engine verbs
// with the tail actions.Dispatch runs after every action (afterLoopAction),
// and record steps the way Dispatch does (loopStep).

// loopGame is a two-seat game at the first seat's precombat main, with
// the first seat holding priority.
func loopGame(t *testing.T) (*Game, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newActiveGame(t)
	advanceIntoStep(t, g, StepPrecombatMain)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	if g.seatIDLocked(g.Turn.PriorityHolder) != me {
		t.Fatalf("setup: priority is with %v, want the active seat", g.seatIDLocked(g.Turn.PriorityHolder))
	}
	return g, me, opp
}

// afterLoopAction is the tail actions.Dispatch runs after an action.
func afterLoopAction(g *Game) {
	g.SettleResolution()
	g.SettleBlockDeclaration()
	g.SettlePassTurn()
	g.TendLoopRecording()
}

// loopStep runs fn as one action by actor, recorded as a step when a
// recording is capturing, in Dispatch's order: draft, act, settle,
// commit, tend.
func loopStep(t *testing.T, g *Game, actor uuid.UUID, typ string, fn func() error) {
	t.Helper()
	d := g.BeginLoopStep()
	if err := fn(); err != nil {
		t.Fatalf("%s by %v: %v", typ, actor, err)
	}
	g.SettleResolution()
	g.SettleBlockDeclaration()
	g.SettlePassTurn()
	g.CommitLoopStep(d, LoopStep{Actor: actor, Player: actor, Type: typ, Label: typ})
	g.TendLoopRecording()
}

// pushLifeAbility puts an ability of controller's on the stack that, as
// it resolves, asks them whether to gain 1 life.
func pushLifeAbility(g *Game, controller uuid.UUID) {
	pushAbilityItem(g, StackItemActivated, controller, uuid.Nil, "gain 1 life?", func(g *Game, it *StackItem) error {
		queueConfirmHooked(g, ConfirmPrompt{Chooser: it.Controller, Question: "Gain 1 life?"}, func(g *Game) error {
			if p := g.playerByIDLocked(it.Controller); p != nil {
				p.Life++
			}
			return nil
		}, nil)
		return nil
	})
}

// loopConfirm is the open yes/no prompt to chooser.
func loopConfirm(t *testing.T, g *Game, chooser uuid.UUID) uuid.UUID {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceConfirm && c.Chooser == chooser {
			return c.ID
		}
	}
	t.Fatalf("no yes/no prompt open for %v (open: %v)", chooser, g.LoopPendingShape())
	return uuid.Nil
}

func loopStatus(g *Game) LoopRecordingStatus {
	if r := g.CurrentLoopRecording(); r != nil {
		return r.Status
	}
	return ""
}

// loopLogLines is every EventLoopRecording line, "<phase>: <label>".
func loopLogLines(g *Game) []string {
	var out []string
	g.WithWriteLock(func() {
		for _, ev := range g.Events {
			if ev.Kind == EventLoopRecording {
				out = append(out, ev.Call+": "+ev.Label)
			}
		}
	})
	return out
}

// playLifeIteration plays one iteration by hand, every action a step:
// the recorder puts the ability on the stack, both seats pass, it
// resolves into the recorder's prompt, and they gain the life.
func playLifeIteration(t *testing.T, g *Game, me, opp uuid.UUID) {
	t.Helper()
	loopStep(t, g, me, "activate_ability", func() error { pushLifeAbility(g, me); return nil })
	loopStep(t, g, me, "pass_priority", g.PassPriority)
	loopStep(t, g, opp, "pass_priority", g.PassPriority)
	if !g.ResolutionPaused() {
		t.Fatal("the ability did not pause on its prompt")
	}
	id := loopConfirm(t, g, me)
	loopStep(t, g, me, "resolve_choice", func() error { return g.ResolveConfirm(id, me, true) })
}

// The owner's report: partway through resolving an ability, the
// recorder asks to record. Before #4736 the server refused
// (ErrLoopPromptOpen); now the request is armed, captures nothing while
// the resolution finishes, and the recording starts at the recorder's
// next priority, where one honest iteration then closes.
func TestLoopRecordingAskedDuringTheRecordersOwnResolutionPrompt(t *testing.T) {
	g, me, opp := loopGame(t)
	pushLifeAbility(g, me)
	resolveUntilPrompt(t, g)
	if !g.ResolutionPaused() {
		t.Fatal("setup: the resolution is not paused on the prompt")
	}

	if err := g.StartLoopRecording(me); err != nil {
		t.Fatalf("StartLoopRecording during a resolution prompt: %v", err)
	}
	afterLoopAction(g)
	if got := loopStatus(g); got != LoopStatusArmed {
		t.Fatalf("status %q while the resolution is paused, want armed", got)
	}
	if d := g.BeginLoopStep(); d != nil {
		t.Fatal("an armed recording drafted a step")
	}

	// Finishing the resolution is not a step: the iteration has not
	// begun. The recording begins at the priority the resolution hands
	// the active player (CR 117.3b).
	id := loopConfirm(t, g, me)
	loopStep(t, g, me, "resolve_choice", func() error { return g.ResolveConfirm(id, me, true) })
	r := g.CurrentLoopRecording()
	if r == nil || r.Status != LoopStatusRecording {
		t.Fatalf("after the resolution finished: %+v, want recording", r)
	}
	if len(r.Steps) != 0 {
		t.Errorf("the answer that finished the resolution was recorded: %+v", r.Steps)
	}
	if r.Start.Priority != me || r.Start.Pending != 0 || len(r.Start.Stack) != 0 {
		t.Errorf("the recording started at %+v, want the recorder's priority with nothing open", r.Start)
	}
	lines := strings.Join(loopLogLines(g), "\n")
	if !strings.Contains(lines, "armed: will record a loop from their next priority") || !strings.Contains(lines, "started: started recording a loop") {
		t.Errorf("log:\n%s", lines)
	}

	// One iteration, which goes through the same kind of prompt, closes.
	playLifeIteration(t, g, me, opp)
	if err := g.StopLoopRecording(me); err != nil {
		t.Fatalf("StopLoopRecording: %v", err)
	}
	r = g.CurrentLoopRecording()
	if r == nil || r.Status != LoopStatusReady {
		t.Fatalf("the iteration did not close: %+v\n%s", r, strings.Join(loopLogLines(g), "\n"))
	}
	if len(r.Steps) != 4 {
		t.Errorf("%d steps, want 4 (activate, two passes, the answer)", len(r.Steps))
	}
	if got := g.LoopNetText(r, 1); !strings.Contains(got, "life +1") {
		t.Errorf("net change %q, want the iteration's life only", got)
	}
}

// Another player's prompt partway through their resolution: the seat
// that asks is allowed to record too, since recording only captures
// what happens next. The other player's answer is not a step, and the
// recording starts where the asker next holds priority.
func TestLoopRecordingAskedDuringAnotherPlayersResolutionPrompt(t *testing.T) {
	g, me, opp := loopGame(t)
	pushLifeAbility(g, opp)
	resolveUntilPrompt(t, g)
	if err := g.StartLoopRecording(me); err != nil {
		t.Fatalf("StartLoopRecording during another player's prompt: %v", err)
	}
	afterLoopAction(g)
	if got := loopStatus(g); got != LoopStatusArmed {
		t.Fatalf("status %q, want armed", got)
	}
	// That player cannot record while one stands.
	if err := g.StartLoopRecording(opp); !errors.Is(err, ErrLoopAlreadyRecording) {
		t.Errorf("a second recording: %v, want ErrLoopAlreadyRecording", err)
	}
	id := loopConfirm(t, g, opp)
	loopStep(t, g, opp, "resolve_choice", func() error { return g.ResolveConfirm(id, opp, false) })
	r := g.CurrentLoopRecording()
	if r == nil || r.Status != LoopStatusRecording || len(r.Steps) != 0 || r.Start.Priority != me {
		t.Fatalf("after their answer: %+v, want an empty recording from my priority", r)
	}
}

// Armed while another seat holds priority with nothing open: nothing is
// captured, so the step may move on, and the recording begins at the
// recorder's next priority, in the step that has.
func TestAnArmedLoopRecordingWaitsForTheRecordersPriority(t *testing.T) {
	g, me, opp := loopGame(t)
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if g.seatIDLocked(g.Turn.PriorityHolder) != opp {
		t.Fatal("setup: the other seat does not hold priority")
	}
	if err := g.StartLoopRecording(me); err != nil {
		t.Fatalf("StartLoopRecording without priority: %v", err)
	}
	afterLoopAction(g)
	if got := loopStatus(g); got != LoopStatusArmed {
		t.Fatalf("status %q, want armed", got)
	}
	main := g.Turn.Step
	loopStep(t, g, opp, "pass_priority", g.PassPriority)
	r := g.CurrentLoopRecording()
	if g.Turn.Step == main {
		t.Fatal("setup: the step did not end")
	}
	if r == nil || r.Status != LoopStatusRecording || len(r.Steps) != 0 || r.Start.Step != g.Turn.Step || r.Start.Priority != me {
		t.Fatalf("after the step moved on: %+v, want an empty recording in step %v", r, g.Turn.Step)
	}
}

// Stop asked without priority, in the middle of the iteration: CR
// 732.2a ends the sequence where a player has priority, so the server
// keeps recording and runs the closure check at the recorder's next
// quiet priority. The steps up to there are the iteration.
func TestLoopStopAskedWithoutPriorityClosesAtTheRecordersNextPriority(t *testing.T) {
	g, me, opp := loopGame(t)
	if err := g.StartLoopRecording(me); err != nil {
		t.Fatal(err)
	}
	if got := loopStatus(g); got != LoopStatusRecording {
		t.Fatalf("status %q with priority and nothing open, want recording at once", got)
	}
	loopStep(t, g, me, "activate_ability", func() error { pushLifeAbility(g, me); return nil })
	loopStep(t, g, me, "pass_priority", g.PassPriority)

	if err := g.StopLoopRecording(me); err != nil {
		t.Fatalf("StopLoopRecording without priority: %v", err)
	}
	r := g.CurrentLoopRecording()
	if r == nil || r.Status != LoopStatusRecording || !r.StopAtPriority {
		t.Fatalf("after stop without priority: %+v, want still recording, stopping", r)
	}
	loopStep(t, g, opp, "pass_priority", g.PassPriority)
	if got := loopStatus(g); got != LoopStatusRecording {
		t.Fatalf("status %q with the resolution paused on the recorder's prompt, want recording", got)
	}
	id := loopConfirm(t, g, me)
	loopStep(t, g, me, "resolve_choice", func() error { return g.ResolveConfirm(id, me, true) })
	r = g.CurrentLoopRecording()
	if r == nil || r.Status != LoopStatusReady || r.StopAtPriority {
		t.Fatalf("at the recorder's priority: %+v, want ready", r)
	}
	if len(r.Steps) != 4 {
		t.Errorf("%d steps, want 4: the steps after the stop are part of the iteration", len(r.Steps))
	}
}

// The closure check still runs at the recorder's priority: stopped with
// the ability still waiting on the stack, the loop does not close, and
// the log names why.
func TestALoopStoppedWithTheStackChangedIsDiscarded(t *testing.T) {
	g, me, _ := loopGame(t)
	if err := g.StartLoopRecording(me); err != nil {
		t.Fatal(err)
	}
	loopStep(t, g, me, "activate_ability", func() error { pushLifeAbility(g, me); return nil })
	if err := g.StopLoopRecording(me); err != nil {
		t.Fatal(err)
	}
	if r := g.CurrentLoopRecording(); r != nil {
		t.Fatalf("a recording that leaves an ability on the stack was kept: %+v", r)
	}
	lines := loopLogLines(g)
	if last := lines[len(lines)-1]; !strings.Contains(last, "the loop does not close: the stack has 1 objects and had 0") {
		t.Errorf("discard line %q", last)
	}
}

// CR 732.2a: "the player with priority may suggest a shortcut". A ready
// loop is refused while the recorder does not hold priority, and while
// the board is not where the iteration started, and is proposed once
// they hold priority there again.
func TestLoopProposalNeedsTheRecordersPriority(t *testing.T) {
	g, me, opp := loopGame(t)
	if err := g.StartLoopRecording(me); err != nil {
		t.Fatal(err)
	}
	playLifeIteration(t, g, me, opp)
	if err := g.StopLoopRecording(me); err != nil {
		t.Fatal(err)
	}
	if got := loopStatus(g); got != LoopStatusReady {
		t.Fatalf("status %q, want ready", got)
	}

	// The recorder passes; the other seat holds priority.
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	afterLoopAction(g)
	if err := g.ProposeLoop(me, 3); !errors.Is(err, ErrLoopNeedsPriority) {
		t.Fatalf("propose without priority: %v, want ErrLoopNeedsPriority", err)
	}
	if got := loopStatus(g); got != LoopStatusReady {
		t.Fatalf("the refused proposal changed the recording: %q", got)
	}
	// They respond; the recorder gets priority over it, and the stack is
	// not the one the loop started from.
	pushLifeAbility(g, opp)
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	afterLoopAction(g)
	if err := g.ProposeLoop(me, 3); !errors.Is(err, ErrLoopBoardChanged) {
		t.Fatalf("propose over a changed stack: %v, want ErrLoopBoardChanged", err)
	}
	// It resolves (their prompt, declined); the recorder has priority on
	// the board the loop started from.
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	afterLoopAction(g)
	id := loopConfirm(t, g, opp)
	if err := g.ResolveConfirm(id, opp, false); err != nil {
		t.Fatal(err)
	}
	afterLoopAction(g)
	if g.seatIDLocked(g.Turn.PriorityHolder) != me {
		t.Fatal("setup: the recorder does not hold priority again")
	}
	if err := g.ProposeLoop(me, 3); err != nil {
		t.Fatalf("propose with priority: %v", err)
	}
	r := g.CurrentLoopRecording()
	if r == nil || r.Status != LoopStatusProposed || r.Proposal.Count != 3 {
		t.Fatalf("after the proposal: %+v", r)
	}
	var asked uuid.UUID
	for _, c := range g.PendingChoices {
		if c.Kind == PendingChoiceLoopProposal {
			asked = c.Chooser
		}
	}
	if asked != opp {
		t.Errorf("CR 732.2b asked %v, want the other seat", asked)
	}
}

// Cancel discards an armed recording as it does any other, and only the
// recorder may; a cancelled request never starts.
func TestAnArmedLoopRecordingCancels(t *testing.T) {
	g, me, opp := loopGame(t)
	pushLifeAbility(g, me)
	resolveUntilPrompt(t, g)
	if err := g.StartLoopRecording(me); err != nil {
		t.Fatal(err)
	}
	if err := g.StopLoopRecording(me); !errors.Is(err, ErrLoopNotRecording) {
		t.Errorf("stop of an armed recording: %v, want ErrLoopNotRecording", err)
	}
	if err := g.CancelLoopRecording(opp); !errors.Is(err, ErrLoopNotRecorder) {
		t.Errorf("another seat cancelled: %v", err)
	}
	if err := g.CancelLoopRecording(me); err != nil {
		t.Fatalf("CancelLoopRecording: %v", err)
	}
	if g.CurrentLoopRecording() != nil {
		t.Fatal("cancel left the recording")
	}
	id := loopConfirm(t, g, me)
	if err := g.ResolveConfirm(id, me, true); err != nil {
		t.Fatal(err)
	}
	afterLoopAction(g)
	if g.CurrentLoopRecording() != nil {
		t.Error("a cancelled request started recording at the next priority")
	}
}

// Who may ask at all: a seat still in a game that is past its opening.
func TestLoopRecordingRefusesOutsideAnActiveGame(t *testing.T) {
	g := newActiveGameMulligansOpen(t, 2)
	if err := g.StartLoopRecording(g.Seats[0].ID); !errors.Is(err, ErrGameNotActive) {
		t.Errorf("during the mulligans: %v, want ErrGameNotActive", err)
	}
	g, _, opp := loopGame(t)
	if err := g.StartLoopRecording(uuid.New()); !errors.Is(err, ErrPlayerNotFound) {
		t.Errorf("not a seat: %v, want ErrPlayerNotFound", err)
	}
	if err := g.Concede(opp); err != nil {
		t.Fatal(err)
	}
	if err := g.StartLoopRecording(opp); err == nil {
		t.Error("a seat that left the game started a recording")
	}
}

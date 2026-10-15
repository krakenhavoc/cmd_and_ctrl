package game

import (
	"testing"

	"github.com/google/uuid"
)

// loop_draw_test.go — CR 732.4 (#2506, ADR 0055's amendment of
// 2026-10-15): "If a loop contains only mandatory actions, the game is
// a draw." The draw is for a loop that repeats unchanged; a loop that
// drains, progresses, grows or holds an optional action is not one.
//
// The loop here is two permanents trading one life back and forth:
// Ebb ("whenever you gain life, you lose 1 life") and Flow ("whenever
// you lose life, you gain 1 life"). Nothing in it targets, nothing in
// it is optional, and every iteration leaves the game exactly where it
// was — the shape CR 104.4b calls a loop "with no way to stop".

const (
	drawOracleEbb  = "test-2506-ebb"
	drawOracleFlow = "test-2506-flow"
	drawLabelEbb   = "Ebb Engine — lose 1 life"
	drawLabelFlow  = "Flow Engine — gain 1 life"
)

// drawLifeTrigger is "whenever <whose> life changes by <sign>, <effect>".
func drawLifeTrigger(label string, applies func(ev Event, source *Card, g *Game) bool, effect func(g *Game, source, controller uuid.UUID) error) TriggeredAbility {
	return TriggeredAbility{
		Watches: []EventKind{EventChangeLife},
		AppliesTo: func(ev Event, source *Card, _ Characteristic, g *Game) bool {
			return applies(ev, source, g)
		},
		Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
			id, controller := source.InstanceID, source.Controller
			return newTriggeredItemForTest(source, label, func(g *Game, _ *StackItem) error {
				return effect(g, id, controller)
			})
		},
	}
}

func drawYouGain(ev Event, source *Card, _ *Game) bool {
	return ev.Target == source.Controller && ev.Amount > 0
}

func drawYouLose(ev Event, source *Card, _ *Game) bool {
	return ev.Target == source.Controller && ev.Amount < 0
}

func drawOpponentLoses(ev Event, source *Card, _ *Game) bool {
	return ev.Target != uuid.Nil && ev.Target != source.Controller && ev.Amount < 0
}

func drawLoseOne(g *Game, source, controller uuid.UUID) error {
	return g.ChangePlayerLifeForEffect(source, controller, -1)
}

func drawGainOne(g *Game, source, controller uuid.UUID) error {
	return g.ChangePlayerLifeForEffect(source, controller, 1)
}

func drawDrainOpponents(g *Game, source, controller uuid.UUID) error {
	for _, p := range g.Seats {
		if p != nil && p.ID != controller && !p.Eliminated {
			if err := g.ChangePlayerLifeForEffect(source, p.ID, -1); err != nil {
				return err
			}
		}
	}
	return nil
}

// seedLifeLoop puts Ebb and Flow on the battlefield under seat 0 with
// the given abilities, then starts the loop with one life gained.
// Returns the two permanents.
func seedLifeLoop(t *testing.T, g *Game, ebb, flow TriggeredAbility) (ebbID, flowID uuid.UUID) {
	t.Helper()
	p := g.Seats[0]
	ebbID, flowID = uuid.New(), uuid.New()
	for _, c := range []Card{
		{InstanceID: ebbID, Name: "Ebb Engine", OracleID: drawOracleEbb, TypeLine: "Enchantment", Owner: p.ID, Controller: p.ID},
		{InstanceID: flowID, Name: "Flow Engine", OracleID: drawOracleFlow, TypeLine: "Enchantment", Owner: p.ID, Controller: p.ID},
	} {
		g.Battlefield.PushTop(c)
	}
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		switch id {
		case drawOracleEbb:
			return []TriggeredAbility{ebb}
		case drawOracleFlow:
			return []TriggeredAbility{flow}
		}
		return nil
	})
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, 1); err != nil {
			t.Fatalf("seed the loop: %v", err)
		}
		g.runStateChecksLocked()
	})
	if len(g.StackMeta) != 1 {
		t.Fatalf("seed: stack items = %d, want 1", len(g.StackMeta))
	}
	return ebbID, flowID
}

// seedPingPong is the mandatory loop that changes nothing.
func seedPingPong(t *testing.T, g *Game) (ebbID, flowID uuid.UUID) {
	t.Helper()
	return seedLifeLoop(t, g,
		drawLifeTrigger(drawLabelEbb, drawYouGain, drawLoseOne),
		drawLifeTrigger(drawLabelFlow, drawYouLose, drawGainOne))
}

// passUntil passes priority one seat at a time until done() or the
// game ends, failing after `limit` passes.
func drawPassUntil(t *testing.T, g *Game, limit int, done func() bool) int {
	t.Helper()
	for i := 0; i < limit; i++ {
		if done() || g.State != StateActive {
			return i
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority after %d passes: %v", i, err)
		}
	}
	if !done() && g.State == StateActive {
		t.Fatalf("still going after %d passes", limit)
	}
	return limit
}

func drawGameOverEvents(g *Game) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventGameOver {
			out = append(out, ev)
		}
	}
	return out
}

// TestAMandatoryLoopThatRepeatsIsADraw is the headline. The breaker
// fires on the loop, sees that its last iteration changed nothing, and
// says so instead of asking for a shortcut. Every seat then passes
// once more with the loop's trigger on the stack, the loop comes round
// unchanged, and the game is a draw.
func TestAMandatoryLoopThatRepeatsIsADraw(t *testing.T) {
	const threshold = 3
	g := newActiveGame(t)
	g.LoopThreshold = threshold
	_, flow := seedPingPong(t, g)

	drawPassUntil(t, g, 200, func() bool { return g.LoopNotice != nil })
	if g.State != StateActive {
		t.Fatalf("the game ended before the table was told: %+v", g.Outcome)
	}
	n := g.CurrentLoopNotice()
	if !n.Mandatory {
		t.Fatalf("notice %+v is not marked mandatory for a loop that changes nothing", n)
	}
	if n.Source != flow {
		t.Errorf("notice names %s, want the Flow Engine (it tripped first)", n.Source)
	}
	if c := loopShortcutPrompt(g); c != nil {
		t.Errorf("a mandatory loop was offered a shortcut: %+v", c)
	}
	if !g.AutoPassSuspended() {
		t.Error("automatic passing is live under a mandatory notice; people must pass by hand")
	}
	if !g.MandatoryLoopStanding() {
		t.Error("MandatoryLoopStanding() = false; bot seats would hold instead of passing")
	}

	// One more time round: Ebb, then Flow, each resolved by every seat
	// passing.
	passes := drawPassUntil(t, g, 50, func() bool { return false })
	if g.State != StateEnded {
		t.Fatalf("game still %s after the loop came round again", g.State)
	}
	if passes != 2*len(g.Seats) {
		t.Errorf("the draw took %d passes after the notice, want %d (two resolutions)", passes, 2*len(g.Seats))
	}
	o := g.Result()
	if o == nil || o.Kind != OutcomeDraw || o.Cause != OutcomeCauseMandatoryLoop || o.Winner != uuid.Nil {
		t.Fatalf("outcome %+v, want a %s draw", o, OutcomeCauseMandatoryLoop)
	}
	if o.Source != flow {
		t.Errorf("outcome source %s, want the Flow Engine %s", o.Source, flow)
	}
	if got := g.OutcomeKind(); got != OutcomeDraw {
		t.Errorf("OutcomeKind() = %q, want %q (the lobby's games.outcome)", got, OutcomeDraw)
	}
	over := drawGameOverEvents(g)
	if len(over) != 1 || over[0].Label != OutcomeCauseMandatoryLoop || over[0].Actor != uuid.Nil || over[0].CardID != flow {
		t.Errorf("game_over events %+v", over)
	}
	if g.LoopNotice != nil {
		t.Errorf("the notice outlived the game: %+v", g.LoopNotice)
	}
	for _, p := range g.Seats {
		if p.Eliminated {
			t.Errorf("%s left the game; a draw leaves everyone in it", p.Name)
		}
	}
	// The resolving ability did nothing more: Flow's gain would have
	// put Ebb's trigger on the stack.
	if len(g.StackMeta) != 0 || len(g.PendingTriggers) != 0 {
		t.Errorf("the ended game still has %d stack items and %d waiting triggers", len(g.StackMeta), len(g.PendingTriggers))
	}
}

// TestADrainingMandatoryLoopIsNotADraw: the same two cards pointed at
// the opponent. Every iteration takes a life total to a new low, so the
// loop is ending the game, and it plays out to the win (CR 704.5a,
// CR 104.2a) without the breaker ever calling it.
func TestADrainingMandatoryLoopIsNotADraw(t *testing.T) {
	g := newActiveGame(t)
	g.LoopThreshold = 3
	me, opp := g.Seats[0], g.Seats[1]
	opp.Life = 12
	seedLifeLoop(t, g,
		drawLifeTrigger(drawLabelEbb, drawYouGain, drawDrainOpponents),
		drawLifeTrigger(drawLabelFlow, drawOpponentLoses, drawGainOne))

	drawPassUntil(t, g, 500, func() bool {
		if g.LoopNotice != nil {
			t.Fatalf("the breaker called a draining loop: %+v", g.LoopNotice)
		}
		return false
	})
	o := g.Result()
	if o == nil || o.Kind != OutcomeWin || o.Winner != me.ID {
		t.Fatalf("outcome %+v, want %s to win", o, me.Name)
	}
}

// TestALoopThatChangesSomethingIsNotADraw: the ping-pong with one
// thing changing each time round — a counter on Flow. The life totals
// make no new lows, so the breaker does call it, but it is not
// repeating: it gets the CR 732 shortcut prompt as before, and a table
// stepping it on by hand never reaches a draw.
func TestALoopThatChangesSomethingIsNotADraw(t *testing.T) {
	g := newActiveGame(t)
	g.LoopThreshold = 3
	seedLifeLoop(t, g,
		drawLifeTrigger(drawLabelEbb, drawYouGain, drawLoseOne),
		drawLifeTrigger(drawLabelFlow, drawYouLose, func(g *Game, source, controller uuid.UUID) error {
			if c := g.findCardByIDLocked(source); c != nil {
				if c.Counters == nil {
					c.Counters = map[string]int{}
				}
				c.Counters["charge"]++
			}
			return drawGainOne(g, source, controller)
		}))

	drawPassUntil(t, g, 200, func() bool { return g.LoopNotice != nil })
	if g.LoopNotice.Mandatory {
		t.Fatalf("notice %+v is mandatory, but every iteration adds a counter", g.LoopNotice)
	}
	if loopShortcutPrompt(g) == nil {
		t.Fatal("no CR 732 shortcut prompt for a loop that is not repeating")
	}
	answerLoopShortcut(t, g, 0)
	for i := 0; i < 10; i++ {
		passAroundOnce(t, g)
		if g.State != StateActive {
			t.Fatalf("stepping a changing loop by hand ended the game: %+v", g.Outcome)
		}
	}
	if g.LoopNotice == nil || g.LoopNotice.Mandatory {
		t.Errorf("notice %+v, want it standing and not mandatory", g.LoopNotice)
	}
}

// TestALoopWithAnOptionalActionIsNotADraw: Flow's gain is a "you may"
// (CR 104.4b: "Loops that contain an optional action don't result in a
// draw"). Answering it is a decision, so the breaker never calls the
// loop, however long the table keeps saying yes.
func TestALoopWithAnOptionalActionIsNotADraw(t *testing.T) {
	g := newActiveGame(t)
	g.LoopThreshold = 3
	flow := drawLifeTrigger(drawLabelFlow, drawYouLose, drawGainOne)
	flow.OptionalPrompt = &TriggerOptionalPrompt{Question: "Gain 1 life?"}
	seedLifeLoop(t, g, drawLifeTrigger(drawLabelEbb, drawYouGain, drawLoseOne), flow)

	answered := 0
	for i := 0; i < 400 && answered < 12; i++ {
		if g.State != StateActive {
			t.Fatalf("the game ended: %+v", g.Outcome)
		}
		if g.LoopNotice != nil {
			t.Fatalf("the breaker called a loop with an optional action: %+v", g.LoopNotice)
		}
		if c := drawFirstChoiceOfKind(g, PendingChoiceTriggerPrompt); c != nil {
			if err := g.ResolveTriggerPrompt(c.ID, c.Chooser, true); err != nil {
				t.Fatalf("ResolveTriggerPrompt: %v", err)
			}
			answered++
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if answered < 12 {
		t.Fatalf("only %d optional answers; the loop stopped turning", answered)
	}
}

func drawFirstChoiceOfKind(g *Game, kind PendingChoiceKind) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind {
			return c
		}
	}
	return nil
}

// TestAStandingAnswerBreaksTheRepeat: a prompt the server answers from
// a standing answer is not a decision for the run (ADR 0127 §7), but it
// is somebody's choice standing in the loop, so the chain of repeated
// positions starts again.
func TestAStandingAnswerBreaksTheRepeat(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.WithWriteLock(func() {
		g.TurnTally.LoopRepeat = map[string]string{"k": "position"}
		g.QueueChoiceForEffect(PendingChoice{Kind: PendingChoiceConfirm, Chooser: p.ID, Reason: "pay?"})
		g.answerChoiceLocked(len(g.PendingChoices)-1, true)
	})
	if g.TurnTally.LoopRepeat != nil {
		t.Errorf("LoopRepeat = %v after an automatic answer, want cleared", g.TurnTally.LoopRepeat)
	}
}

// TestAStepThroughNoticeTurnsMandatoryOnceTheLoopSettles: a notice
// raised before the loop settled (here, a stopped shortcut) is marked
// mandatory the first time the loop comes round unchanged, and the
// draw needs one more unchanged iteration after that, so the table has
// seen the banner before passing into it.
func TestAStepThroughNoticeTurnsMandatoryOnceTheLoopSettles(t *testing.T) {
	g := newActiveGame(t)
	g.LoopThreshold = 3
	_, flow := seedPingPong(t, g)
	drawPassUntil(t, g, 200, func() bool { return g.LoopNotice != nil })
	// Pretend the notice went up before the loop settled.
	g.WithWriteLock(func() {
		g.LoopNotice.Mandatory = false
		g.TurnTally.LoopRepeat = nil
	})
	passAroundOnce(t, g) // Ebb
	passAroundOnce(t, g) // Flow: no position to compare with yet
	if g.State != StateActive || g.LoopNotice == nil || g.LoopNotice.Mandatory {
		t.Fatalf("state %s, notice %+v: want active and not yet mandatory", g.State, g.LoopNotice)
	}
	passAroundOnce(t, g) // Ebb
	passAroundOnce(t, g) // Flow: the same position again
	if g.State != StateActive || g.LoopNotice == nil || !g.LoopNotice.Mandatory {
		t.Fatalf("state %s, notice %+v: want active and now mandatory", g.State, g.LoopNotice)
	}
	passAroundOnce(t, g)
	passAroundOnce(t, g)
	if o := g.Result(); o == nil || o.Cause != OutcomeCauseMandatoryLoop || o.Source != flow {
		t.Fatalf("outcome %+v, want the mandatory loop's draw", o)
	}
}

// TestADecisionUnderAMandatoryNoticeIsNotADraw: anybody may break the
// loop (CR 732.5 only says nobody has to). A decision clears the
// notice and every run, so the loop has to be found again from the
// start before it can be a draw.
func TestADecisionUnderAMandatoryNoticeIsNotADraw(t *testing.T) {
	g := newActiveGame(t)
	g.LoopThreshold = 3
	seedPingPong(t, g)
	drawPassUntil(t, g, 200, func() bool { return g.LoopNotice != nil })
	g.WithWriteLock(func() { g.EmitEvent(Event{Kind: EventCast, Actor: g.Seats[1].ID, CardID: uuid.New()}) })
	if g.LoopNotice != nil {
		t.Fatalf("a cast left the notice standing: %+v", g.LoopNotice)
	}
	passAroundOnce(t, g)
	passAroundOnce(t, g)
	if g.State != StateActive {
		t.Fatalf("the loop drew the game right after a player acted: %+v", g.Outcome)
	}
}

// TestALoopReplayNeverDraws: ADR 0193 §5's replay is a shortcut the
// table agreed to. The breaker sits out, and so does the draw.
func TestALoopReplayNeverDraws(t *testing.T) {
	g := newActiveGame(t)
	g.LoopThreshold = 3
	seedPingPong(t, g)
	g.WithWriteLock(func() { g.loopReplaying = true })
	for i := 0; i < 20; i++ {
		passAroundOnce(t, g)
	}
	if g.State != StateActive || g.LoopNotice != nil {
		t.Fatalf("state %s, notice %+v during a replay", g.State, g.LoopNotice)
	}
	if len(g.TurnTally.LoopRepeat) != 0 {
		t.Errorf("positions taken during a replay: %v", g.TurnTally.LoopRepeat)
	}
}

// TestLoopRepeatSurvivesCloneAndSnapshot: the positions ride the tally
// both ways, and so does the notice's mandatory mark.
func TestLoopRepeatSurvivesCloneAndSnapshot(t *testing.T) {
	g := newActiveGame(t)
	g.LoopThreshold = 3
	seedPingPong(t, g)
	drawPassUntil(t, g, 200, func() bool { return g.LoopNotice != nil })
	want := len(g.TurnTally.LoopRepeat)
	if want == 0 || !g.LoopNotice.Mandatory {
		t.Fatalf("setup: positions %v, notice %+v", g.TurnTally.LoopRepeat, g.LoopNotice)
	}
	clone := g.Clone()
	g.WithWriteLock(func() { g.TurnTally.LoopRepeat["x"] = "y" })
	if len(clone.TurnTally.LoopRepeat) != want {
		t.Errorf("the clone shares LoopRepeat with the live game")
	}
	restored, err := clone.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if len(restored.TurnTally.LoopRepeat) != want || restored.LoopNotice == nil || !restored.LoopNotice.Mandatory {
		t.Errorf("restored positions %v, notice %+v", restored.TurnTally.LoopRepeat, restored.LoopNotice)
	}
}

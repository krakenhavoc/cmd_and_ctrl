package game

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/uuid"
)

// turn_control_test.go — ADR 0167 PR 1: a turn-long PlayerControl
// record (CR 723.1, Emrakul, the Promised End), driven through
// BeginTurnControlForEffect and the rotation seam. Emrakul is not
// registered yet, so every test schedules the control directly.

const emrakulName = "Emrakul, the Promised End"

// scheduleTurnControl makes seat `controller` control seat `controlled`
// during that seat's next turn, as Emrakul's cast trigger would.
func scheduleTurnControl(t *testing.T, g *Game, controller, controlled int) uuid.UUID {
	t.Helper()
	src := uuid.New()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.BeginTurnControlForEffect(TurnControl{
			Controller: g.Seats[controller].ID,
			Controlled: g.Seats[controlled].ID,
			Source:     src,
			SourceName: emrakulName,
		})
	})
	if !ok {
		t.Fatalf("BeginTurnControlForEffect(%d controls %d) refused", controller, controlled)
	}
	return src
}

func turnControlEvents(g *Game, kind EventKind) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == kind {
			out = append(out, ev)
		}
	}
	return out
}

func controllerOfSeat(g *Game, seat int) int {
	id := g.ControllerOfPlayer(g.Seats[seat].ID)
	for i, p := range g.Seats {
		if p.ID == id {
			return i
		}
	}
	return -1
}

// The record waits, applies for the whole of the controlled player's
// next turn, ends as the turn after it begins, and gives them an extra
// turn directly after it.
func TestTurnControlCoversTheirNextTurnAndOwesAnExtraTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: seat %d starts, want 0", g.Turn.ActiveSeat)
	}
	src := scheduleTurnControl(t, g, 0, 1)

	if got := controllerOfSeat(g, 1); got != 1 {
		t.Fatalf("a waiting record applied: seat 1 is controlled by seat %d", got)
	}
	var sched []PlayerControl
	g.ReadSnapshot(func() { sched = g.ScheduledPlayerControlsLocked() })
	if len(sched) != 1 || sched[0].SourceName != emrakulName {
		t.Fatalf("scheduled controls = %+v, want the one record", sched)
	}
	if n := len(turnControlEvents(g, EventPlayerControlScheduled)); n != 1 {
		t.Errorf("EventPlayerControlScheduled emitted %d times, want 1", n)
	}

	passTurn(t, g) // seat 1's turn: controlled
	if g.Turn.ActiveSeat != 1 || g.Turn.Extra {
		t.Fatalf("after seat 0's turn: %+v, want seat 1's normal turn", shapeOf(g))
	}
	if got := controllerOfSeat(g, 1); got != 0 {
		t.Fatalf("seat 1's turn is controlled by seat %d, want seat 0", got)
	}
	if got := g.ControlUntilOfPlayer(g.Seats[1].ID); got != ControlUntilTurn {
		t.Errorf("ControlUntilOfPlayer = %q, want %q", got, ControlUntilTurn)
	}
	// CR 723.3: still the active player.
	if g.Seats[g.Turn.ActiveSeat].ID != g.Seats[1].ID {
		t.Error("the controlled player is no longer the active player")
	}
	controlledSeq := g.Turn.Seq
	if g.PlayerControls[0].TurnSeq != controlledSeq || g.PlayerControls[0].Stage != ControlTurn {
		t.Errorf("record = %+v, want stage turn on Seq %d", g.PlayerControls[0], controlledSeq)
	}
	// The control begins before the turn's own boundary event, so the
	// untap step's choices are already the controller's.
	began := turnControlEvents(g, EventPlayerControlBegan)
	if len(began) != 1 || began[0].Actor != g.Seats[0].ID || began[0].Target != g.Seats[1].ID || began[0].Source != src {
		t.Fatalf("EventPlayerControlBegan = %+v", began)
	}
	for _, ev := range turnControlEvents(g, EventTurnBegan) {
		if ev.Amount == controlledSeq && ev.Seq < began[0].Seq {
			t.Error("the turn began before its control did")
		}
	}
	if len(g.ScheduledPlayerControls()) != 0 {
		t.Error("a begun record is still listed as scheduled")
	}

	passTurn(t, g) // seat 1's extra turn: not controlled
	if g.Turn.ActiveSeat != 1 || !g.Turn.Extra {
		t.Fatalf("after the controlled turn: %+v, want seat 1's extra turn", shapeOf(g))
	}
	if got := controllerOfSeat(g, 1); got != 1 {
		t.Errorf("seat 1's extra turn is controlled by seat %d; the control ended with the turn", got)
	}
	if len(g.PlayerControls) != 0 {
		t.Errorf("records left after the controlled turn: %+v", g.PlayerControls)
	}
	if n := len(turnControlEvents(g, EventPlayerControlEnded)); n != 1 {
		t.Errorf("EventPlayerControlEnded emitted %d times, want 1", n)
	}
	added := turnControlEvents(g, EventExtraTurnAdded)
	if len(added) != 1 || added[0].Actor != g.Seats[1].ID || added[0].Source != src || added[0].Amount != g.Turn.ExtraRef {
		t.Errorf("EventExtraTurnAdded = %+v, want one for seat 1's extra turn (ref %d)", added, g.Turn.ExtraRef)
	}

	passTurn(t, g)
	if got := shapeOf(g); got != (turnShape{seat: 2, round: 1}) {
		t.Errorf("after the extra turn: %+v, want seat 2's normal turn", got)
	}
}

// ScheduledPlayerControls is ScheduledPlayerControlsLocked under the
// read lock, for these tests.
func (g *Game) ScheduledPlayerControls() []PlayerControl {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.ScheduledPlayerControlsLocked()
}

// A record created during the controlled player's own turn waits for
// their next one: only a turn that begins later can pick it up.
func TestTurnControlCreatedDuringTheirTurnWaitsForTheNext(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleTurnControl(t, g, 2, 0)
	if got := controllerOfSeat(g, 0); got != 0 {
		t.Fatalf("seat 0's current turn became controlled by seat %d", got)
	}
	for seat := 1; seat <= 3; seat++ {
		passTurn(t, g)
		if got := controllerOfSeat(g, 0); got != 0 {
			t.Fatalf("seat %d's turn: seat 0 controlled by seat %d", seat, got)
		}
	}
	passTurn(t, g)
	if g.Turn.ActiveSeat != 0 || controllerOfSeat(g, 0) != 2 {
		t.Fatalf("seat 0's next turn: active %d, controlled by seat %d; want seat 0, by seat 2",
			g.Turn.ActiveSeat, controllerOfSeat(g, 0))
	}
}

// CR 723.1a: the last effect created is the one that works. Each
// Emrakul effect still owes its own extra turn (ruling).
func TestTwoTurnControlsNewestWinsAndEachOwesAnExtraTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleTurnControl(t, g, 2, 1)
	scheduleTurnControl(t, g, 3, 1)
	if n := len(g.ScheduledPlayerControls()); n != 2 {
		t.Fatalf("%d scheduled, want 2: a second Emrakul deletes nothing", n)
	}
	passTurn(t, g)
	if got := controllerOfSeat(g, 1); got != 3 {
		t.Fatalf("seat 1 is controlled by seat %d, want seat 3 (the newest effect)", got)
	}
	var live []PlayerControl
	g.ReadSnapshot(func() { live = g.LivePlayerControlsLocked() })
	if len(live) != 1 || live[0].Controller != g.Seats[3].ID {
		t.Errorf("live records = %+v, want only seat 3's", live)
	}
	passTurn(t, g)
	if q := g.ExtraTurnsQueued(); len(q) != 1 {
		t.Errorf("queue after the first extra turn began: %+v, want one more turn", q)
	}
	for i := 0; i < 2; i++ {
		if g.Turn.ActiveSeat != 1 || !g.Turn.Extra {
			t.Fatalf("extra turn %d: %+v, want seat 1's extra turn", i+1, shapeOf(g))
		}
		if got := controllerOfSeat(g, 1); got != 1 {
			t.Errorf("extra turn %d is controlled by seat %d", i+1, got)
		}
		passTurn(t, g)
	}
	if got := shapeOf(g); got != (turnShape{seat: 2, round: 1}) {
		t.Errorf("after both extra turns: %+v, want seat 2's normal turn", got)
	}
}

// CR 500.7: the extra turn goes directly after the controlled turn,
// below every extra turn created during that turn (also "after this
// one", and created later) and above every turn queued before it.
func TestTurnControlExtraTurnGoesBelowTurnsCreatedDuringIt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	// Seat 1 has two extra turns queued (Time Stretch); the first of
	// them is the controlled one.
	stretch := takeExtraTurns(t, g, 1, 2)
	src := scheduleTurnControl(t, g, 0, 1)
	var emrakulRef int
	g.ReadSnapshot(func() { emrakulRef = g.PlayerControls[0].ExtraTurn.Ref })

	passTurn(t, g)
	if g.Turn.ActiveSeat != 1 || g.Turn.ExtraRef != stretch[0] || controllerOfSeat(g, 1) != 0 {
		t.Fatalf("Time Stretch's first turn: %+v ref %d, controlled by %d", shapeOf(g), g.Turn.ExtraRef, controllerOfSeat(g, 1))
	}
	// During the controlled turn, seat 3 is given an extra turn.
	warp := takeExtraTurns(t, g, 3, 1)

	want := []struct {
		seat, ref int
	}{
		{3, warp[0]},    // created during the controlled turn: first
		{1, emrakulRef}, // then "after that turn"
		{1, stretch[1]}, // then the turn queued before it
	}
	for i, w := range want {
		passTurn(t, g)
		if g.Turn.ActiveSeat != w.seat || !g.Turn.Extra || g.Turn.ExtraRef != w.ref {
			t.Fatalf("turn %d: seat %d extra=%v ref %d; want seat %d's extra turn ref %d",
				i+1, g.Turn.ActiveSeat, g.Turn.Extra, g.Turn.ExtraRef, w.seat, w.ref)
		}
	}
	for _, ev := range turnControlEvents(g, EventExtraTurnAdded) {
		if ev.Amount == emrakulRef && ev.Source != src {
			t.Errorf("the anchored turn's event names source %v, want %v", ev.Source, src)
		}
	}
	passTurn(t, g)
	if got := shapeOf(g); got != (turnShape{seat: 1, round: 1}) {
		t.Errorf("after the extra turns: %+v, want seat 1's normal turn (rotation resumes from seat 0)", got)
	}
}

// CR 800.4a: the controller leaving ends the control, and the extra turn
// is still owed (it is a separate part of the effect).
func TestTurnControlControllerLeavesDuringTheTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleTurnControl(t, g, 2, 1)
	passTurn(t, g)
	if controllerOfSeat(g, 1) != 2 {
		t.Fatal("setup: seat 1's turn is not controlled")
	}
	if err := g.Concede(g.Seats[2].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if got := controllerOfSeat(g, 1); got != 1 {
		t.Fatalf("seat 1 is still controlled by seat %d after the controller left", got)
	}
	passTurn(t, g)
	if g.Turn.ActiveSeat != 1 || !g.Turn.Extra {
		t.Errorf("after the turn: %+v, want seat 1's extra turn", shapeOf(g))
	}
	if n := len(turnControlEvents(g, EventPlayerControlEnded)); n != 0 {
		t.Errorf("an ended line for a control that had already ended (%d)", n)
	}
}

// The controller leaving before the turn: the turn begins uncontrolled,
// and the extra turn is still owed.
func TestTurnControlControllerLeavesBeforeTheTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleTurnControl(t, g, 3, 1)
	if err := g.Concede(g.Seats[3].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if n := len(g.ScheduledPlayerControls()); n != 0 {
		t.Errorf("a control whose controller left is still scheduled in the view (%d)", n)
	}
	passTurn(t, g)
	if g.Turn.ActiveSeat != 1 || controllerOfSeat(g, 1) != 1 {
		t.Fatalf("seat 1's turn: active %d, controlled by %d", g.Turn.ActiveSeat, controllerOfSeat(g, 1))
	}
	if n := len(turnControlEvents(g, EventPlayerControlBegan)); n != 0 {
		t.Errorf("a began line for a controller who left (%d)", n)
	}
	passTurn(t, g)
	if g.Turn.ActiveSeat != 1 || !g.Turn.Extra {
		t.Errorf("after the turn: %+v, want seat 1's extra turn", shapeOf(g))
	}
}

// The controlled player leaving before the turn: it never begins, so
// there is neither control nor an extra turn, and the record goes.
func TestTurnControlControlledLeavesBeforeTheTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleTurnControl(t, g, 0, 1)
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	g.WithWriteLock(g.prunePlayerControlsLocked)
	if len(g.PlayerControls) != 0 {
		t.Errorf("records left: %+v", g.PlayerControls)
	}
	for i := 0; i < 3; i++ {
		passTurn(t, g)
		if g.Turn.ActiveSeat == 1 || g.Turn.Extra {
			t.Fatalf("turn %d: %+v; seat 1 left and is owed nothing", i+1, shapeOf(g))
		}
	}
	if n := len(turnControlEvents(g, EventExtraTurnAdded)); n != 0 {
		t.Errorf("%d extra turns queued for a player who left", n)
	}
}

// The controlled player leaving during the turn: the turn ends without
// them (CR 800.4j, ADR 0059's simplification) and the extra turn never
// begins (ruling, CR 800.4k).
func TestTurnControlControlledLeavesDuringTheTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleTurnControl(t, g, 0, 1)
	passTurn(t, g)
	if controllerOfSeat(g, 1) != 0 {
		t.Fatal("setup: seat 1's turn is not controlled")
	}
	// CR 723.6: the controlled player may concede.
	if err := g.Concede(g.Seats[1].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	if got := shapeOf(g); got != (turnShape{seat: 2, round: 1}) {
		t.Errorf("after the controlled player left: %+v, want seat 2's normal turn", got)
	}
	if len(g.ExtraTurns) != 0 || len(turnControlEvents(g, EventExtraTurnAdded)) != 0 {
		t.Errorf("an extra turn was queued for a player who left: %+v", g.ExtraTurns)
	}
	if len(g.PlayerControls) != 0 {
		t.Errorf("records left: %+v", g.PlayerControls)
	}
}

// Word of Command cast during an Emrakul turn controls the player only
// while it resolves; the turn's record applies again afterwards.
func TestWordOfCommandInsideAnEmrakulTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleTurnControl(t, g, 0, 1)
	passTurn(t, g)
	bob := g.Seats[1]
	if len(bob.Hand.Cards) == 0 {
		t.Fatal("setup: seat 1 has no hand")
	}
	card := bob.Hand.Cards[0].InstanceID
	var prompt uuid.UUID
	g.WithWriteLock(func() {
		prompt = g.BeginControlledPlayForEffect(ControlledPlay{
			Controller: g.Seats[2].ID,
			Controlled: bob.ID,
			Source:     uuid.New(),
			SourceName: "Word of Command",
			Card:       card,
		})
	})
	if prompt == uuid.Nil {
		t.Fatal("no controlled_play prompt")
	}
	if len(g.PlayerControls) != 2 {
		t.Fatalf("records = %+v; Word of Command must not delete the turn's record", g.PlayerControls)
	}
	if got := controllerOfSeat(g, 1); got != 2 {
		t.Fatalf("during Word of Command seat 1 is controlled by seat %d, want seat 2", got)
	}
	if got := g.ControlUntilOfPlayer(bob.ID); got != ControlUntilResolution {
		t.Errorf("ControlUntilOfPlayer = %q, want %q", got, ControlUntilResolution)
	}
	// "It can't be played" (this package has no probe, so it is taken).
	if err := g.ResolvePendingChoice(prompt, g.Seats[2].ID, nil); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if got := controllerOfSeat(g, 1); got != 0 {
		t.Fatalf("after Word of Command seat 1 is controlled by seat %d, want seat 0 again", got)
	}
	if got := g.ControlUntilOfPlayer(bob.ID); got != ControlUntilTurn {
		t.Errorf("ControlUntilOfPlayer = %q, want %q", got, ControlUntilTurn)
	}
}

// A standing answer never answers a controlled chooser's prompt: the
// rule is theirs, and the decision is not (ADR 0167 §4).
func TestStandingAnswerSkipsAControlledChooser(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleTurnControl(t, g, 0, 1)
	passTurn(t, g)
	bob := g.Seats[1]
	c := &PendingChoice{ID: uuid.New(), Chooser: bob.ID, AutoAnswerKey: "k"}
	var reason AskedByHand
	g.WithWriteLock(func() { reason = g.autoAnswerHandReasonLocked(c, AutoAnswerNever) })
	if reason != AskedByHandControlled {
		t.Errorf("hand reason = %q, want %q", reason, AskedByHandControlled)
	}
	g.WithWriteLock(func() { reason = g.autoAnswerHandReasonLocked(&PendingChoice{Chooser: g.Seats[2].ID}, AutoAnswerNever) })
	if reason != "" {
		t.Errorf("an uncontrolled chooser's hand reason = %q, want none", reason)
	}
}

// The record is plain data: it survives a snapshot round trip in both
// stages, and a clone never shares its extra turn with the game.
func TestTurnControlSnapshotRoundTrip(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	scheduleTurnControl(t, g, 0, 1)
	scheduleTurnControl(t, g, 3, 2)
	passTurn(t, g)
	clone := g.Clone()

	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	back, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !reflect.DeepEqual(back.PlayerControls, g.PlayerControls) {
		t.Fatalf("restored records\n %+v\nwant\n %+v", back.PlayerControls, g.PlayerControls)
	}
	if back.PlayerControls[0].Stage != ControlTurn || back.PlayerControls[1].Stage != ControlAwaitingTurn {
		t.Fatalf("stages = %q, %q", back.PlayerControls[0].Stage, back.PlayerControls[1].Stage)
	}
	if controllerOfSeat(back, 1) != 0 {
		t.Error("the restored game lost the live control")
	}
	if back.PlayerControls[0].ExtraTurn == g.PlayerControls[0].ExtraTurn {
		t.Error("the restored record shares its extra turn with the original")
	}
	if clone.PlayerControls[1].ExtraTurn == g.PlayerControls[1].ExtraTurn {
		t.Error("a clone shares its extra turn with the original")
	}
	// The rest of the life of the restored record: it ends, and its
	// extra turn is taken.
	passTurn(t, back)
	if back.Turn.ActiveSeat != 1 || !back.Turn.Extra {
		t.Errorf("restored game after the controlled turn: %+v, want seat 1's extra turn", shapeOf(back))
	}
	// The clone was untouched by the original's turn ending.
	passTurn(t, g)
	if clone.PlayerControls[0].Stage != ControlTurn || len(clone.PlayerControls) != 2 {
		t.Errorf("the clone changed with the game: %+v", clone.PlayerControls)
	}
}

// A record written before ADR 0167 has no Until, and reads as a Word of
// Command record.
func TestPlayerControlWithoutUntilReadsAsResolution(t *testing.T) {
	if got := (PlayerControl{}).UntilKind(); got != ControlUntilResolution {
		t.Errorf("UntilKind of an old record = %q, want %q", got, ControlUntilResolution)
	}
}

// BeginTurnControlForEffect refuses a player out of the game and a
// player controlling themselves.
func TestBeginTurnControlRefusals(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	if err := g.Concede(g.Seats[3].ID); err != nil {
		t.Fatalf("Concede: %v", err)
	}
	for _, tc := range []TurnControl{
		{Controller: g.Seats[0].ID, Controlled: g.Seats[3].ID},
		{Controller: g.Seats[3].ID, Controlled: g.Seats[1].ID},
		{Controller: g.Seats[1].ID, Controlled: g.Seats[1].ID},
	} {
		var ok bool
		g.WithWriteLock(func() { ok = g.BeginTurnControlForEffect(tc) })
		if ok {
			t.Errorf("BeginTurnControlForEffect(%+v) was accepted", tc)
		}
	}
	if len(g.PlayerControls) != 0 {
		t.Errorf("records = %+v", g.PlayerControls)
	}
}

// ADR 0167, 2026-10-16 amendment: a plain "you control target player
// during that player's next turn" (Mindslaver, Worst Fears, Sorin
// Markov's -7) controls the same turn and owes no extra one.
func TestTurnControlWithoutAnExtraTurnOwesNone(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	var ok bool
	g.WithWriteLock(func() {
		ok = g.BeginTurnControlForEffect(TurnControl{
			Controller:  g.Seats[0].ID,
			Controlled:  g.Seats[1].ID,
			Source:      uuid.New(),
			SourceName:  "Mindslaver",
			NoExtraTurn: true,
		})
	})
	if !ok {
		t.Fatal("BeginTurnControlForEffect refused")
	}
	if g.NextExtraRef != 0 {
		t.Errorf("NextExtraRef = %d: a control with no extra turn minted a ref", g.NextExtraRef)
	}
	passTurn(t, g) // seat 1's turn: controlled
	if got := controllerOfSeat(g, 1); got != 0 {
		t.Fatalf("seat 1's turn is controlled by seat %d, want seat 0", got)
	}
	passTurn(t, g) // seat 2's turn: no extra turn for seat 1
	if got := shapeOf(g); got != (turnShape{seat: 2, round: 1}) {
		t.Errorf("after the controlled turn: %+v, want seat 2's normal turn", got)
	}
	if len(g.PlayerControls) != 0 {
		t.Errorf("records left after the controlled turn: %+v", g.PlayerControls)
	}
	if n := len(turnControlEvents(g, EventExtraTurnAdded)); n != 0 {
		t.Errorf("EventExtraTurnAdded emitted %d times, want 0", n)
	}
	if n := len(turnControlEvents(g, EventPlayerControlEnded)); n != 1 {
		t.Errorf("EventPlayerControlEnded emitted %d times, want 1", n)
	}
}

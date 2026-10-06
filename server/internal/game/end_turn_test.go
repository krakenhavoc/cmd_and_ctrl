package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// end_turn_test.go — CR 724.1, "end the turn" (#2165). The resolver is
// stubbed (effect_hooks_test.go's withEffectHooks), so these tests
// drive the engine verb through a real resolution without importing
// the catalog. The catalog's card tests are
// effects/end_the_turn_cards_test.go; the enumerator's are
// legal/end_turn_test.go.

const endTurnStubOracle = "stub-end-the-turn"

// withEndTurnResolver installs a resolver under which a spell with
// endTurnStubOracle runs `before` (if any) and then ends the turn — the
// shape of Ultima, whose wipe comes first.
func withEndTurnResolver(t *testing.T, before func(g *Game, item *StackItem)) {
	t.Helper()
	withEffectHooks(t,
		func(g *Game, item *StackItem, oracleID string) error {
			if oracleID != endTurnStubOracle {
				return nil
			}
			if before != nil {
				before(g, item)
			}
			g.EndTheTurnForEffect(item.SourceCardID)
			return nil
		},
		nil,
		func(oracleID string) bool { return oracleID == endTurnStubOracle },
	)
}

// pushEndTurnSpell puts an instant carrying the stub oracle in p's hand.
func pushEndTurnSpell(p *Player) uuid.UUID {
	id := pushTypedCardToHand(p, "Time Stop Stub", "Instant")
	for i := range p.Hand.Cards {
		if p.Hand.Cards[i].InstanceID == id {
			p.Hand.Cards[i].OracleID = endTurnStubOracle
		}
	}
	return id
}

// castEndTurnSpell casts a fresh stub from caster's hand.
func castEndTurnSpell(t *testing.T, g *Game, caster *Player) uuid.UUID {
	t.Helper()
	id := pushEndTurnSpell(caster)
	if err := g.CastSpell(caster.ID, id, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	return id
}

// stepsBegunInTurn lists the steps that began in turn `seq`, in order.
func stepsBegunInTurn(g *Game, seq int) []Step {
	var out []Step
	for _, ev := range g.Events {
		if ev.Kind == EventStepBegan && ev.Amount == seq {
			out = append(out, ev.Step)
		}
	}
	return out
}

// eventsSince returns the events of a kind emitted after seq.
func eventsSince(g *Game, kind EventKind, seq uint64) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Seq > seq && ev.Kind == kind {
			out = append(out, ev)
		}
	}
	return out
}

// The whole of CR 724.1 in one resolution, cast in the declare blockers
// step: every object on the stack is exiled — the resolving spell, a
// spell beneath it, an ability beneath that — and none of it counts as
// countered; every creature leaves combat with no combat damage dealt;
// nothing between the declare blockers step and cleanup begins; and
// with nothing waiting in the cleanup step the turn ends and the next
// player's begins.
func TestEndTheTurnExilesTheStackRemovesCombatAndSkipsToCleanup(t *testing.T) {
	withEndTurnResolver(t, nil)
	g := newActiveGame(t)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	blocker := pushCombatant(t, g, g.Seats[1], "Blocker", 1, 1)
	blockAfterLockIn(t, g, attacker, blocker)

	active := g.Seats[0]
	under := pushTypedCardToHand(active, "Lesser Instant", "Instant")
	if err := g.CastSpell(active.ID, under, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	abilityRan := false
	g.WithWriteLock(func() {
		id := uuid.New()
		g.StackMeta[id] = &StackItem{
			ID: id, Kind: StackItemActivated, Controller: active.ID, Owner: active.ID,
			SourceCardID: attacker, Label: "an ability beneath", Seq: g.nextStackSeqLocked(),
			Effect: func(*Game, *StackItem) error {
				abilityRan = true
				return nil
			},
		}
	})
	stop := castEndTurnSpell(t, g, active)
	turn := g.Turn.Seq
	seq := lastSeq(g)
	passUntilStackEmpty(t, g)

	if g.Turn.Seq != turn+1 || g.Turn.ActiveSeat != 1 {
		t.Fatalf("the turn did not end: turn %d seat %d step %s", g.Turn.Seq, g.Turn.ActiveSeat, g.Turn.Step)
	}
	if g.Turn.Step != StepUpkeep || g.Turn.PriorityHolder != 1 {
		t.Errorf("next turn at %s with priority %d, want upkeep with seat 1", g.Turn.Step, g.Turn.PriorityHolder)
	}
	if g.Turn.PassedInSuccession != 0 {
		t.Errorf("the ended turn's passes carried into the next turn (CR 117.4): %v", g.Turn.PassedInSuccession.Seats())
	}
	for _, id := range []uuid.UUID{stop, under} {
		if !inExile(g, id) {
			t.Errorf("%s was not exiled (CR 724.1b)", id)
		}
	}
	if abilityRan || len(g.StackMeta) != 0 || len(g.Stack.Cards) != 0 {
		t.Errorf("the stack was not exiled: ability ran=%v, %d records, %d cards", abilityRan, len(g.StackMeta), len(g.Stack.Cards))
	}
	if n := len(eventsSince(g, EventCounterSpell, seq)); n != 0 {
		t.Errorf("%d stack objects were reported countered; exiling is not countering", n)
	}
	for _, ev := range eventsSince(g, EventDealDamage, seq) {
		if ev.Combat {
			t.Errorf("combat damage after the turn ended: %+v", ev)
		}
	}
	for _, id := range []uuid.UUID{attacker, blocker} {
		c := findCard(g, id)
		if c == nil {
			t.Fatalf("%s died; no combat damage should have been dealt", id)
		}
		if c.AttackingTarget != uuid.Nil || c.BlockingTarget != uuid.Nil {
			t.Errorf("%s is still in combat", c.Name)
		}
	}
	// #1501's declaration record goes with the combat it described, so
	// the next combat's declare blockers step starts with nobody done.
	if len(g.blockedAttackers) != 0 || len(g.blocksDeclared) != 0 || len(g.announcedBlocks) != 0 {
		t.Errorf("the block declaration record survived the ended turn: blocked %v declared %v announced %d",
			g.blockedAttackers, g.blocksDeclared, len(g.announcedBlocks))
	}
	got := stepsBegunInTurn(g, turn)
	if last := got[len(got)-1]; last != StepCleanup {
		t.Errorf("the ended turn's last step was %s, want cleanup", last)
	}
	for _, s := range got {
		switch s {
		case StepCombatDamage, StepEndCombat, StepPostcombatMain, StepEnd:
			t.Errorf("%s began after the turn ended (CR 724.1d); steps: %v", s, got)
		}
	}
	ended := eventsSince(g, EventTurnEnded, seq)
	if len(ended) != 1 {
		t.Fatalf("%d EventTurnEnded, want 1", len(ended))
	}
	if ev := ended[0]; ev.Actor != active.ID || ev.Source != stop || ev.Amount != turn {
		t.Errorf("EventTurnEnded = %+v, want actor %s source %s amount %d", ev, active.ID, stop, turn)
	}
	if g.TurnEndPending {
		t.Errorf("TurnEndPending was left set")
	}
}

// CR 724.1e: the end step never begins, so neither an "at the beginning
// of your end step" trigger nor a delayed "at the beginning of the next
// end step" one fires in the turn that ended. The delayed one waits for
// the next end step that does begin — the next turn's.
func TestEndTheTurnSkipsEndStepTriggersAndDelayedOnesWait(t *testing.T) {
	withEndTurnResolver(t, nil)
	g := newActiveGame(t)
	active := g.Seats[0]
	const watcherOracle = "test-end-step-watcher"
	g.Battlefield.PushTop(Card{
		InstanceID: uuid.New(), Name: "End Step Watcher", OracleID: watcherOracle,
		TypeLine: "Enchantment", Owner: active.ID, Controller: active.ID,
	})
	endStepFired := 0
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != watcherOracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventBeginEndStep},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				endStepFired++
				return newTriggeredItemForTest(source, "at the beginning of your end step", nil)
			},
		}}
	})
	advanceTo(t, g, StepPrecombatMain)
	delayedFired := 0
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller: active.ID,
			Label:      "at the beginning of the next end step",
			At:         StepEnd,
			Body: testBody(func(*Game, *StackItem) error {
				delayedFired++
				return nil
			}),
		})
	})
	castEndTurnSpell(t, g, active)
	turn := g.Turn.Seq
	passUntilStackEmpty(t, g)

	if g.Turn.Seq != turn+1 {
		t.Fatalf("the turn did not end")
	}
	if endStepFired != 0 || delayedFired != 0 {
		t.Fatalf("end-step triggers fired in the ended turn: watcher %d, delayed %d (CR 724.1e)", endStepFired, delayedFired)
	}
	queued := 0
	for _, dt := range g.DelayedTriggers {
		if dt != nil && dt.At == StepEnd {
			queued++
		}
	}
	if queued != 1 {
		t.Fatalf("the next-end-step trigger was dropped (%d queued); it waits for the next end step", queued)
	}

	// The next turn's end step fires it — on the opponent's turn, since
	// it is not "your next end step" — and only it: the watcher reads
	// its controller's own end step.
	advanceTo(t, g, StepEnd)
	passUntilStackEmpty(t, g)
	if delayedFired != 1 {
		t.Errorf("the delayed trigger fired %d times in the next turn's end step, want 1", delayedFired)
	}
	if endStepFired != 0 {
		t.Errorf("the controller's end-step trigger fired on the opponent's turn")
	}
}

// CR 724.1 + 514.1 / 514.2: in the cleanup step the turn skipped to,
// the active player discards to hand size, marked damage wears off and
// "until end of turn" effects end. While the discard is owed nobody
// holds priority — the resolution does not hand it back — and the
// discard finishes the turn.
func TestEndTheTurnCleanupDiscardsClearsDamageAndEndsUntilEndOfTurn(t *testing.T) {
	withEndTurnResolver(t, nil)
	g := newActiveGame(t)
	active := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	bear := pushCombatant(t, g, active, "Bear", 2, 2)
	g.WithWriteLock(func() {
		findCard(g, bear).DamageMarked = 1
		g.RegisterScopedEffectForEffect(uuid.New(), g.PinnedObjectsLocked(bear),
			[]Mod{ModifyPTMod(3, 3)}, g.UntilEndOfTurnDuration(), "test — +3/+3 until end of turn")
		g.RecomputeLayersIfStaleLocked()
	})
	if p := findCard(g, bear).CurrentPower(); p != 5 {
		t.Fatalf("setup: the pump is not on (power %d)", p)
	}
	stop := pushEndTurnSpell(active)
	fillHandTo(t, g, active, 10)
	if err := g.CastSpell(active.ID, stop, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	turn := g.Turn.Seq
	passUntilStackEmpty(t, g)

	if g.Turn.Seq != turn || g.Turn.Step != StepCleanup {
		t.Fatalf("at turn %d %s, want this turn's cleanup step", g.Turn.Seq, g.Turn.Step)
	}
	if g.DiscardPending[active.ID] != 2 {
		t.Fatalf("discard pending = %v, want 2 for the active player (CR 514.1)", g.DiscardPending)
	}
	if g.Turn.PriorityHolder != NoPriority {
		t.Errorf("seat %d holds priority in a cleanup step waiting on a discard", g.Turn.PriorityHolder)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := findCard(g, bear)
	if c.DamageMarked != 0 {
		t.Errorf("marked damage %d survived the cleanup step (CR 514.2)", c.DamageMarked)
	}
	if p := c.CurrentPower(); p != 2 {
		t.Errorf("power %d, want 2: the until-end-of-turn pump should have ended (CR 514.2)", p)
	}

	if err := g.DiscardSelection(active.ID, []uuid.UUID{active.Hand.Cards[0].InstanceID, active.Hand.Cards[1].InstanceID}); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}
	if g.Turn.Seq != turn+1 {
		t.Errorf("the discard did not finish the turn: turn %d step %s", g.Turn.Seq, g.Turn.Step)
	}
}

// CR 724.1f / 514.3a: a trigger the cleanup step causes (here, from its
// own hand-size discard) goes on the stack in that cleanup step, the
// active player gets priority there, and once the table has passed with
// the stack empty another cleanup step begins before the turn ends.
func TestEndTheTurnCleanupTriggerGetsPriorityAndAnotherCleanupStep(t *testing.T) {
	withEndTurnResolver(t, nil)
	g := newActiveGame(t)
	active := g.Seats[0]
	installDiscardLifeWatcher(t, g, active)
	advanceTo(t, g, StepPrecombatMain)
	stop := pushEndTurnSpell(active)
	fillHandTo(t, g, active, 9)
	if err := g.CastSpell(active.ID, stop, CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	turn := g.Turn.Seq
	life := active.Life
	passUntilStackEmpty(t, g)
	if g.DiscardPending[active.ID] != 1 {
		t.Fatalf("discard pending = %v, want 1", g.DiscardPending)
	}
	if err := g.DiscardSelection(active.ID, []uuid.UUID{active.Hand.Cards[0].InstanceID}); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}
	if g.Turn.Step != StepCleanup || g.Turn.Seq != turn {
		t.Fatalf("the turn ended past a waiting trigger: turn %d step %s", g.Turn.Seq, g.Turn.Step)
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat || len(g.StackMeta) != 1 {
		t.Fatalf("CR 514.3a: want the trigger on the stack and the active player on priority; holder %d, %d items",
			g.Turn.PriorityHolder, len(g.StackMeta))
	}
	// CR 117.4: the passes made before the turn ended (the ones that
	// resolved the spell) do not count toward this window.
	if g.Turn.PassedInSuccession != 0 {
		t.Errorf("the succession carried into the cleanup window: %v", g.Turn.PassedInSuccession.Seats())
	}
	passUntilStackEmpty(t, g)
	if active.Life != life+3 {
		t.Errorf("life %d, want %d: the discard trigger should have resolved in the ended turn", active.Life, life+3)
	}
	closeCleanupWindow(t, g)
	if g.Turn.Seq != turn+1 {
		t.Fatalf("the turn did not end after the second cleanup step: turn %d step %s", g.Turn.Seq, g.Turn.Step)
	}
	n := 0
	for _, s := range stepsBegunInTurn(g, turn) {
		if s == StepCleanup {
			n++
		}
	}
	if n != 2 {
		t.Errorf("%d cleanup steps began in the ended turn, want 2 (CR 514.3a)", n)
	}
}

// installDiscardLifeWatcher puts a permanent under p that gains its
// controller 3 life whenever a card is discarded.
func installDiscardLifeWatcher(t *testing.T, g *Game, p *Player) {
	t.Helper()
	const oracle = "test-end-turn-discard-watcher"
	g.Battlefield.PushTop(Card{
		InstanceID: uuid.New(), Name: "Discard Watcher", OracleID: oracle,
		TypeLine: "Enchantment", Owner: p.ID, Controller: p.ID,
	})
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches:   []EventKind{EventDiscardCard},
			AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true },
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return newTriggeredItemForTest(source, "Discard Watcher — gain 3 life",
					func(g *Game, item *StackItem) error {
						return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 3)
					})
			},
		}}
	})
}

// CR 724.1c: state-based actions are checked BEFORE the cleanup step,
// so a creature the ending effect left lethally damaged dies — the
// cleanup step's damage removal would otherwise have saved it — and its
// dies trigger, which nobody may put on the stack yet, goes on it in
// the cleanup step with the active player getting priority (724.1f).
func TestEndTheTurnChecksStateBasedActionsBeforeCleanup(t *testing.T) {
	g := newActiveGame(t)
	active := g.Seats[0]
	const oracle = "test-end-turn-dies-watcher"
	doomed := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: doomed, Name: "Doomed", OracleID: oracle, TypeLine: "Creature — Test",
		Power: 2, Toughness: 2, Owner: active.ID, Controller: active.ID,
	})
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventLTB},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return newTriggeredItemForTest(source, "Doomed — gain 3 life",
					func(g *Game, item *StackItem) error {
						return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, 3)
					})
			},
		}}
	})
	withEndTurnResolver(t, func(g *Game, _ *StackItem) {
		findCard(g, doomed).DamageMarked = 2
	})
	advanceTo(t, g, StepPrecombatMain)
	castEndTurnSpell(t, g, active)
	turn := g.Turn.Seq
	life := active.Life
	passUntilStackEmpty(t, g)

	if findCard(g, doomed) != nil {
		t.Fatalf("the lethally damaged creature survived: the CR 724.1c check did not run before cleanup's damage removal")
	}
	if g.Turn.Seq != turn || g.Turn.Step != StepCleanup {
		t.Fatalf("at turn %d %s, want this turn's cleanup step with the dies trigger waiting", g.Turn.Seq, g.Turn.Step)
	}
	if active.Life != life+3 {
		t.Errorf("life %d, want %d: the dies trigger should have resolved in the cleanup step (CR 724.1f)", active.Life, life+3)
	}
	closeCleanupWindow(t, g)
	if g.Turn.Seq != turn+1 {
		t.Errorf("the turn did not end after the cleanup window")
	}
}

// CR 724.1a: a trigger that triggered before the process began and is
// not on the stack yet ceases to exist — including one still being
// announced (its "you may" open).
func TestEndTheTurnDropsTriggersThatHaveNotReachedTheStack(t *testing.T) {
	ran := false
	withEndTurnResolver(t, func(g *Game, item *StackItem) {
		g.PendingTriggers = append(g.PendingTriggers, &StackItem{
			ID: uuid.New(), Kind: StackItemTriggered, Controller: item.Controller,
			Label: "pending before the process",
			Effect: func(*Game, *StackItem) error {
				ran = true
				return nil
			},
		})
		g.QueueChoiceForEffect(PendingChoice{
			Kind: PendingChoiceTriggerOrder, Chooser: item.Controller,
			TriggerOrderIDs: []uuid.UUID{uuid.New(), uuid.New()},
		})
	})
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	castEndTurnSpell(t, g, g.Seats[0])
	turn := g.Turn.Seq
	passUntilStackEmpty(t, g)

	if ran {
		t.Errorf("a trigger that had not reached the stack resolved (CR 724.1a)")
	}
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceTriggerOrder {
			t.Errorf("the trigger-order prompt survived the process")
		}
	}
	if g.Turn.Seq != turn+1 {
		t.Errorf("the turn did not end: turn %d step %s", g.Turn.Seq, g.Turn.Step)
	}
}

// Ending the turn from a CR 514.3a priority window: the cleanup step in
// progress ends and another begins, and the trigger that opened the
// window is exiled with the stack instead of resolving.
func TestEndTheTurnFromACleanupPriorityWindow(t *testing.T) {
	withEndTurnResolver(t, nil)
	g := newActiveGame(t)
	active := g.Seats[0]
	installDiscardLifeWatcher(t, g, active)
	fillHandTo(t, g, active, 8)
	advanceIntoCleanup(t, g)
	if err := g.DiscardSelection(active.ID, []uuid.UUID{active.Hand.Cards[0].InstanceID}); err != nil {
		t.Fatalf("DiscardSelection: %v", err)
	}
	if g.Turn.Step != StepCleanup || g.Turn.PriorityHolder != g.Turn.ActiveSeat || len(g.StackMeta) != 1 {
		t.Fatalf("setup: no CR 514.3a window (step %s holder %d items %d)", g.Turn.Step, g.Turn.PriorityHolder, len(g.StackMeta))
	}
	turn := g.Turn.Seq
	life := active.Life
	stop := castEndTurnSpell(t, g, active)
	passUntilStackEmpty(t, g)

	if !inExile(g, stop) {
		t.Errorf("the spell was not exiled")
	}
	if active.Life != life {
		t.Errorf("the trigger under the spell resolved (life %d → %d); it was exiled with the stack", life, active.Life)
	}
	if g.Turn.Seq != turn+1 {
		t.Fatalf("the turn did not end: turn %d step %s", g.Turn.Seq, g.Turn.Step)
	}
	n := 0
	for _, s := range stepsBegunInTurn(g, turn) {
		if s == StepCleanup {
			n++
		}
	}
	if n != 2 {
		t.Errorf("%d cleanup steps began, want 2: the window's and the one the turn skipped to", n)
	}
}

// The verb called from a prompt's continuation after its resolution has
// already closed — no resolution is open to settle — still gets its
// boundary from SettleResolution, which the dispatcher runs after every
// action.
func TestEndTheTurnOutsideAResolutionSettlesAtTheNextAction(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	turn := g.Turn.Seq
	g.WithWriteLock(func() { g.EndTheTurnForEffect(uuid.Nil) })
	if !g.TurnEndPending || g.Turn.Step != StepPrecombatMain || g.Turn.PriorityHolder != NoPriority {
		t.Fatalf("pending=%v at %s holder %d, want the rest of the process owed with priority parked",
			g.TurnEndPending, g.Turn.Step, g.Turn.PriorityHolder)
	}
	g.SettleResolution()
	if g.TurnEndPending || g.Turn.Seq != turn+1 {
		t.Errorf("SettleResolution did not finish the ended turn: pending=%v turn %d %s",
			g.TurnEndPending, g.Turn.Seq, g.Turn.Step)
	}
}

// AdvanceStep's CR 117.4 drive resolves the stack before the cursor
// moves, and a resolution that ends the turn ends the drive there.
func TestEndTheTurnThroughAdvanceStep(t *testing.T) {
	withEndTurnResolver(t, nil)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	castEndTurnSpell(t, g, g.Seats[0])
	turn := g.Turn.Seq
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if g.Turn.Seq != turn+1 || g.Turn.Step != StepUpkeep {
		t.Errorf("at turn %d %s, want the next turn's upkeep", g.Turn.Seq, g.Turn.Step)
	}
	for _, s := range stepsBegunInTurn(g, turn) {
		if s == StepPostcombatMain || s == StepEnd || s == StepBeginCombat {
			t.Errorf("%s began in the ended turn", s)
		}
	}
}

// The deferred half rides Clone, the snapshot and SettleResolution, and
// an undo to before the resolution replays it.
func TestEndTheTurnPendingHalfSurvivesCloneSnapshotAndUndo(t *testing.T) {
	withEndTurnResolver(t, nil)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	castEndTurnSpell(t, g, g.Seats[0])
	turn := g.Turn.Seq
	pre := g.Clone()

	// The resolution alone, with no boundary after it.
	resolveTop(t, g)
	// 724.1a and 724.1b have happened; 724.1c onwards waits for the
	// boundary, so the cursor has not moved and nobody holds priority.
	if !g.TurnEndPending || g.Turn.Step != StepPrecombatMain || g.Turn.Seq != turn {
		t.Fatalf("after the resolution: pending=%v at turn %d %s, want pending in this turn's main phase",
			g.TurnEndPending, g.Turn.Seq, g.Turn.Step)
	}
	if g.Turn.PriorityHolder != NoPriority {
		t.Errorf("seat %d holds priority between the resolution and its boundary (CR 724.1f)", g.Turn.PriorityHolder)
	}
	if len(g.Stack.Cards) != 0 {
		t.Errorf("the stack was not exiled inside the resolution (CR 724.1b)")
	}
	if n := countEvents(g, EventStepBegan) - countEvents(pre, EventStepBegan); n != 0 {
		t.Errorf("%d steps began inside the resolution; the cleanup step waits for the boundary", n)
	}
	if !g.Clone().TurnEndPending {
		t.Errorf("Clone dropped TurnEndPending")
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var back GameSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if !back.TurnEndPending {
		t.Errorf("the snapshot dropped TurnEndPending")
	}

	g.SettleResolution()
	if g.TurnEndPending || g.Turn.Seq != turn+1 {
		t.Fatalf("SettleResolution did not finish the turn: pending=%v turn %d %s", g.TurnEndPending, g.Turn.Seq, g.Turn.Step)
	}

	// Undo to before the resolution: the spell is back on the stack and
	// the turn is in its main phase; resolving it again ends it again.
	g.WithWriteLock(func() { g.RestoreFrom(pre) })
	if g.Turn.Seq != turn || g.Turn.Step != StepPrecombatMain || g.TurnEndPending || len(g.Stack.Cards) != 1 {
		t.Fatalf("undo did not restore the main phase with the spell on the stack: turn %d %s pending=%v stack %d",
			g.Turn.Seq, g.Turn.Step, g.TurnEndPending, len(g.Stack.Cards))
	}
	passUntilStackEmpty(t, g)
	if g.Turn.Seq != turn+1 {
		t.Errorf("replaying after the undo did not end the turn")
	}
}

// A commander spell exiled from the stack by the process is offered
// the command zone (CR 903.9a, ADR 0115) as part of the 724.1c check.
// The cleanup step waits on the answer rather than walking past it,
// and the turn still ends once it is given.
func TestEndTheTurnExilingACommanderSpellAsksAndThenEnds(t *testing.T) {
	withEndTurnResolver(t, nil)
	g := newActiveGame(t)
	active := g.Seats[0]
	advanceTo(t, g, StepPrecombatMain)
	cmdr := NewCommander("Test Commander", active.ID)
	cmdr.TypeLine = "Legendary Creature — Bear"
	cmdr.Power, cmdr.Toughness = 2, 2
	cmdr.Controller = active.ID
	active.Hand.PushTop(cmdr)
	if err := g.CastSpell(active.ID, cmdr.InstanceID, CastSpellParams{}); err != nil {
		t.Fatalf("cast the commander: %v", err)
	}
	castEndTurnSpell(t, g, active)
	turn := g.Turn.Seq
	passUntilStackEmpty(t, g)

	prompts := commanderReturnPrompts(g)
	if len(prompts) != 1 {
		t.Fatalf("%d commander_return prompts, want 1 (CR 903.9a); step %s turn %d", len(prompts), g.Turn.Step, g.Turn.Seq)
	}
	if g.Turn.Seq != turn {
		t.Fatalf("the turn ended past an open commander question")
	}
	if g.Turn.PriorityHolder != NoPriority {
		t.Errorf("seat %d holds priority while the question is open", g.Turn.PriorityHolder)
	}
	if err := g.ResolveCommanderReturn(prompts[0].ID, active.ID, true); err != nil {
		t.Fatalf("ResolveCommanderReturn: %v", err)
	}
	if !active.Command.Contains(cmdr.InstanceID) {
		t.Errorf("the commander is not in the command zone")
	}
	for i := 0; i < 4 && g.Turn.Seq == turn && g.Turn.Step == StepCleanup && g.Turn.PriorityHolder != NoPriority; i++ {
		closeCleanupWindow(t, g)
	}
	if g.Turn.Seq != turn+1 {
		t.Errorf("the turn did not end after the answer: turn %d %s holder %d", g.Turn.Seq, g.Turn.Step, g.Turn.PriorityHolder)
	}
}

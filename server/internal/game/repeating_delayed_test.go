package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// repeating_delayed_test.go — #2169, CR 603.7b: a delayed trigger with a
// stated duration ("whenever … this turn") triggers every time its event
// happens until the duration ends, and its CR 605.1b triggered-mana twin.

// scheduleRepeatingOnCast is scheduleOnCast with Repeats set. The source
// is a card that does not exist: the spell that made the trigger is long
// gone, and the trigger must not care.
func scheduleRepeatingOnCast(g *Game, controller uuid.UUID, fired *int) {
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller:   controller,
			SourceCardID: uuid.New(),
			Label:        "probe — whenever you cast this turn",
			On:           []EventKind{EventCast},
			Repeats:      true,
			Condition: testCondition(func(ev Event, dt *DelayedTrigger, _ *Game) bool {
				return ev.Actor == dt.Controller
			}),
			Body: testBody(func(_ *Game, _ *StackItem) error {
				*fired++
				return nil
			}),
		})
	})
}

func TestRepeatingDelayedTriggerFiresOnEveryEventThisTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	fired := 0
	scheduleRepeatingOnCast(g, me.ID, &fired)

	emitCast(g, me.ID, uuid.New())
	settleStack(t, g)
	emitCast(g, me.ID, uuid.New())
	settleStack(t, g)
	if fired != 2 {
		t.Fatalf("effect ran %d times after two casts, want 2 (CR 603.7b)", fired)
	}
	if len(g.DelayedTriggers) != 1 {
		t.Fatalf("%d delayed triggers queued, want it to stay until the turn ends", len(g.DelayedTriggers))
	}
	// A cast by someone else is not its event and does not end it.
	emitCast(g, g.Seats[1].ID, uuid.New())
	settleStack(t, g)
	if fired != 2 || len(g.DelayedTriggers) != 1 {
		t.Errorf("a non-matching cast changed things: fired %d, queued %d", fired, len(g.DelayedTriggers))
	}
}

func TestRepeatingDelayedTriggerEndsWithTheTurn(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	fired := 0
	scheduleRepeatingOnCast(g, me.ID, &fired)
	emitCast(g, me.ID, uuid.New())
	settleStack(t, g)

	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	if len(g.DelayedTriggers) != 0 {
		t.Fatalf("%d triggers survived cleanup, want 0 (CR 514.2)", len(g.DelayedTriggers))
	}
	emitCast(g, me.ID, uuid.New())
	settleStack(t, g)
	if fired != 1 {
		t.Errorf("fired %d times, want 1: next turn's cast owes nothing", fired)
	}
}

// The trigger is the spell's, not a permanent's: scheduling it from a
// source that is not on the battlefield (or anywhere) still fires it,
// and the trigger leaves no trace on a permanent.
func TestRepeatingDelayedTriggerSurvivesItsSourceLeaving(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := pushBattlefieldForTest(g, me.ID, "Source", "Artifact", "repeat-source")
	fired := 0
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller:   me.ID,
			SourceCardID: src,
			Label:        "probe",
			On:           []EventKind{EventCast},
			Repeats:      true,
			Body:         testBody(func(_ *Game, _ *StackItem) error { fired++; return nil }),
		})
		if i := findCardOnBattlefield(g, src); i >= 0 {
			g.Battlefield.Cards = append(g.Battlefield.Cards[:i:i], g.Battlefield.Cards[i+1:]...)
		}
	})
	emitCast(g, me.ID, uuid.New())
	settleStack(t, g)
	emitCast(g, me.ID, uuid.New())
	settleStack(t, g)
	if fired != 2 {
		t.Fatalf("fired %d times after its source left, want 2", fired)
	}
}

func TestRepeatingDelayedTriggerSnapshotRoundTrip(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	fired := 0
	scheduleRepeatingOnCast(g, me.ID, &fired)
	emitCast(g, me.ID, uuid.New())
	settleStack(t, g)

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a game owing a repeating delayed trigger is not a restore point: %+v", snap.Continuations)
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		t.Fatal(err)
	}
	var back GameSnapshot
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	restored, err := back.RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	if len(restored.DelayedTriggers) != 1 || !restored.DelayedTriggers[0].Repeats {
		t.Fatalf("restored queue = %+v, want one trigger that still repeats", restored.DelayedTriggers)
	}
	emitCast(restored, me.ID, uuid.New())
	settleStack(t, restored)
	emitCast(restored, me.ID, uuid.New())
	settleStack(t, restored)
	if fired != 3 {
		t.Errorf("fired %d times in all, want 1 before the snapshot and 2 after", fired)
	}

	// Undo: a clone rewinds to a game that still owes it.
	cl := g.Clone()
	if len(cl.DelayedTriggers) != 1 || !cl.DelayedTriggers[0].Repeats {
		t.Errorf("Clone lost Repeats: %+v", cl.DelayedTriggers)
	}
}

// A trigger doubler doubles a delayed trigger like any other: one cast
// puts two items on the stack, and the second cast two more.
func TestTriggerDoublerDoublesARepeatingDelayedTrigger(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const doublerOracle = "repeat-doubler"
	pushBattlefieldForTest(g, me.ID, "Doubler", "Artifact", doublerOracle)
	withCatalogTriggerDoublers(t, func(key string) []TriggerDoubler {
		if key != doublerOracle {
			return nil
		}
		return []TriggerDoubler{{Label: "Doubler", Applies: func(*Game, TriggerDoublingQuery) bool { return true }}}
	})
	fired := 0
	scheduleRepeatingOnCast(g, me.ID, &fired)

	emitCast(g, me.ID, uuid.New())
	if len(g.PendingTriggers) != 2 {
		t.Fatalf("one event queued %d triggers, want 2 with a doubler", len(g.PendingTriggers))
	}
	settleStack(t, g)
	emitCast(g, me.ID, uuid.New())
	settleStack(t, g)
	if fired != 4 {
		t.Errorf("fired %d times over two events, want 4", fired)
	}
}

// --- the triggered-mana twin (CR 605.1b) ------------------------------

func TestTurnManaTriggerAddsForAnyPlayersSwampThisTurn(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
			Controller:     me.ID,
			SourceCardID:   uuid.New(),
			Label:          "Bubbling Muck — add an additional {B}",
			ManaTapSubtype: "Swamp",
			ManaAdds:       "{B}",
			Repeats:        true,
		})
	})
	mine := pushBattlefieldForTest(g, me.ID, "Swamp", "Basic Land — Swamp", "test-swamp")
	theirs := pushBattlefieldForTest(g, opp.ID, "Swamp", "Basic Land — Swamp", "test-swamp")
	forest := pushForest(g, me)

	if err := g.ActivateManaAbility(me.ID, mine, 0, ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolColors(me); got["B"] != 2 {
		t.Errorf("my pool = %v, want the Swamp's {B} and the extra", got)
	}
	if err := g.ActivateManaAbility(opp.ID, theirs, 0, ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolColors(opp); got["B"] != 2 {
		t.Errorf("their pool = %v, want a player's own Swamp to add the extra too", got)
	}
	if err := g.ActivateManaAbility(me.ID, forest, 0, ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := poolColors(me); got["G"] != 1 || got["B"] != 2 {
		t.Errorf("my pool after a Forest = %v, want no extra for a non-Swamp", got)
	}
	if len(g.PendingTriggers) != 0 || len(g.StackMeta) != 0 {
		t.Errorf("a mana trigger used the stack: %d pending, %d meta", len(g.PendingTriggers), len(g.StackMeta))
	}

	snap := g.CaptureSnapshot()
	if !snap.Restorable() {
		t.Fatalf("a game owing a turn mana trigger is not a restore point: %+v", snap.Continuations)
	}
	restored, err := snap.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	if len(restored.DelayedTriggers) != 1 || restored.DelayedTriggers[0].ManaAdds != "{B}" {
		t.Errorf("restored queue = %+v", restored.DelayedTriggers)
	}

	advanceToStepOfSeat(t, g, 1, StepUpkeep)
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("the mana trigger survived cleanup: %+v", g.DelayedTriggers)
	}
}

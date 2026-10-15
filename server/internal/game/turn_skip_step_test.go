package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// turn_skip_step_test.go pins ADR 0188 Decision 3 (#3563): a step of one
// extra turn skipped (ModSkipExtraTurnStep) and rules effects queued with
// an extra turn (ExtraTurn.Effects). The card-level behaviours (Savor the
// Moment, Alchemist's Gambit) are in cards/effects/savor_gambit_test.go;
// these drive the engine seam directly.

func skipStepOf(t *testing.T, g *Game, ref int, step Step) {
	t.Helper()
	ok := false
	g.WithWriteLock(func() {
		ok = g.SkipStepOfExtraTurnForEffect(ref, step, uuid.Nil, "probe — skip the step")
	})
	if !ok {
		t.Fatalf("SkipStepOfExtraTurnForEffect(%d, %s) wrote nothing", ref, step)
	}
}

func duringTurn(t *testing.T, g *Game, ref int) {
	t.Helper()
	ok := false
	g.WithWriteLock(func() {
		ok = g.DuringExtraTurnRuleEffectForEffect(ref, uuid.Nil, ScopeGame, uuid.Nil,
			[]Mod{{Kind: ModDamageCantBePrevented}}, "probe — during that turn")
	})
	if !ok {
		t.Fatalf("DuringExtraTurnRuleEffectForEffect(%d) wrote nothing", ref)
	}
}

func countSkipStepRecords(g *Game) int {
	n := 0
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == ModSkipExtraTurnStep {
				n++
			}
		}
	}
	return n
}

func pushTappedRockForSkipTest(g *Game, seat int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: id, Name: "Tapped Rock", TypeLine: "Artifact",
		Owner: g.Seats[seat].ID, Controller: g.Seats[seat].ID, Tapped: true,
	})
	return id
}

func rockTappedNow(g *Game, id uuid.UUID) bool {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return g.Battlefield.Cards[i].Tapped
		}
	}
	return false
}

// TestAStepSkippedOfOneExtraTurnIsPassedOverOnce: the untap step of the
// bound turn is walked past, nothing untaps, the cursor lands on the
// upkeep, and the record is spent.
func TestAStepSkippedOfOneExtraTurnIsPassedOverOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	rock := pushTappedRockForSkipTest(g, 0)
	refs := takeExtraTurns(t, g, 0, 1)
	skipStepOf(t, g, refs[0], StepUntap)

	passTurn(t, g)
	if got := shapeOf(g); got != (turnShape{seat: 0, extra: true, round: 1}) {
		t.Fatalf("after the turn: %+v, want seat 0's extra turn", got)
	}
	if g.Turn.Step != StepUpkeep {
		t.Errorf("the extra turn starts in %s, want the upkeep", g.Turn.Step)
	}
	if !rockTappedNow(g, rock) {
		t.Error("the permanent untapped although the untap step was skipped")
	}
	if n := countSkipStepRecords(g); n != 0 {
		t.Errorf("the spent skip left %d records", n)
	}
	if got := countEvents(g, EventStepBegan); got == 0 {
		t.Error("no step began at all")
	}
}

// TestASkippedStepIsBoundToItsOwnTurn: two extra turns are queued and
// the skip names the one taken SECOND (the stack takes the most recent
// first). The first one untaps normally; the second skips.
func TestASkippedStepIsBoundToItsOwnTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	rock := pushTappedRockForSkipTest(g, 0)
	refs := takeExtraTurns(t, g, 0, 2) // refs[0] is taken first
	skipStepOf(t, g, refs[1], StepUntap)

	passTurn(t, g)
	if g.Turn.ExtraRef != refs[0] {
		t.Fatalf("first turn taken has ref %d, want %d", g.Turn.ExtraRef, refs[0])
	}
	if rockTappedNow(g, rock) {
		t.Error("the first extra turn's untap step was skipped by the second turn's record")
	}
	if n := countSkipStepRecords(g); n != 1 {
		t.Fatalf("%d records after the first extra turn, want the second turn's 1", n)
	}
	// Tap it again; the second turn must not untap it.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == rock {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	passTurn(t, g)
	if g.Turn.ExtraRef != refs[1] {
		t.Fatalf("second turn taken has ref %d, want %d", g.Turn.ExtraRef, refs[1])
	}
	if !rockTappedNow(g, rock) {
		t.Error("the second extra turn's untap step was not skipped")
	}
	if n := countSkipStepRecords(g); n != 0 {
		t.Errorf("%d records left after the second turn", n)
	}
}

// TestASkippedStepOfASkippedTurnLeavesNoRecord: a skip debt skips the
// extra turn, and the record bound to it goes with it (CR 614.10a).
func TestASkippedStepOfASkippedTurnLeavesNoRecord(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	refs := takeExtraTurns(t, g, 1, 1)
	skipStepOf(t, g, refs[0], StepUntap)
	duringTurn(t, g, refs[0])
	skipTurns(t, g, 1, uuid.Nil, 1)

	passTurn(t, g)
	if got := shapeOf(g); got.extra {
		t.Fatalf("after the turn: %+v, want the extra turn skipped", got)
	}
	if n := countSkipStepRecords(g); n != 0 {
		t.Errorf("a skipped turn left %d skipExtraTurnStep records", n)
	}
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == ModDamageCantBePrevented {
				t.Error("a skipped turn started its during-that-turn effect")
			}
		}
	}
	if len(g.ExtraTurns) != 0 {
		t.Errorf("the queue still holds %d turns", len(g.ExtraTurns))
	}
}

// TestSkipStepOfExtraTurnRefusesWhatItCannotBind: no queued turn with
// that ref, the cleanup step, an unknown step.
func TestSkipStepOfExtraTurnRefusesWhatItCannotBind(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	refs := takeExtraTurns(t, g, 0, 1)
	for name, c := range map[string]struct {
		ref  int
		step Step
	}{
		"unqueued ref": {ref: refs[0] + 7, step: StepUntap},
		"cleanup":      {ref: refs[0], step: StepCleanup},
		"unknown step": {ref: refs[0], step: Step("sleep")},
	} {
		ok := true
		g.WithWriteLock(func() { ok = g.SkipStepOfExtraTurnForEffect(c.ref, c.step, uuid.Nil, "probe") })
		if ok {
			t.Errorf("%s: accepted", name)
		}
	}
	if n := countSkipStepRecords(g); n != 0 {
		t.Errorf("refused calls wrote %d records", n)
	}
}

// TestSkipExtraTurnStepModProblem pins the registration and restore
// check, including the fields on another kind.
func TestSkipExtraTurnStepModProblem(t *testing.T) {
	for name, m := range map[string]Mod{
		"no turn":         {Kind: ModSkipExtraTurnStep, Step: StepUntap},
		"no step":         {Kind: ModSkipExtraTurnStep, ExtraTurn: 3},
		"cleanup":         {Kind: ModSkipExtraTurnStep, Step: StepCleanup, ExtraTurn: 3},
		"step on another": {Kind: ModSkipTurns, Player: uuid.New(), Amount: 1, Step: StepUntap},
		"ref on another":  {Kind: ModAttackTax, ExtraTurn: 2},
	} {
		if skipExtraTurnStepModProblem(m) == "" {
			t.Errorf("%s: accepted", name)
		}
	}
	if p := skipExtraTurnStepModProblem(Mod{Kind: ModSkipExtraTurnStep, Step: StepUntap, ExtraTurn: 3}); p != "" {
		t.Errorf("a sound mod refused: %s", p)
	}
	if p := skipExtraTurnStepModProblem(Mod{Kind: ModPreventDamage, Amount: 3}); p != "" {
		t.Errorf("another kind refused: %s", p)
	}
}

// TestAnEffectQueuedWithAnExtraTurnStartsWithIt: before the turn begins
// nothing is registered, as it begins the rules effect is, with the
// timestamp it was created with, lasting until the end of that turn.
func TestAnEffectQueuedWithAnExtraTurnStartsWithIt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	refs := takeExtraTurns(t, g, 0, 1)
	duringTurn(t, g, refs[0])
	if len(g.ScopedEffects) != 0 {
		t.Fatalf("the effect began in the turn that queued it: %+v", g.ScopedEffects)
	}
	if len(g.DamageCantBePreventedThisTurnLabels()) != 0 {
		t.Fatal("the turn that queued the effect already shows it")
	}
	stamp := g.ExtraTurns[0].Effects[0].Timestamp
	if stamp == 0 {
		t.Fatal("the queued effect carries no timestamp")
	}

	passTurn(t, g)
	if got := shapeOf(g); !got.extra {
		t.Fatalf("after the turn: %+v, want the extra turn", got)
	}
	if len(g.ScopedEffects) != 1 {
		t.Fatalf("%d records in the extra turn, want 1", len(g.ScopedEffects))
	}
	rec := g.ScopedEffects[0]
	if rec.Timestamp != stamp {
		t.Errorf("registered timestamp %d, want the creation timestamp %d (CR 613.7a)", rec.Timestamp, stamp)
	}
	if rec.Duration.Kind != UntilEndOfTurn || rec.Duration.Player != g.Seats[0].ID {
		t.Errorf("duration %+v, want until end of the extra turn", rec.Duration)
	}
	if rec.Seq == 0 {
		t.Error("a rules record is named, and got no Seq")
	}
	if len(g.DamageCantBePreventedThisTurnLabels()) != 1 {
		t.Error("the extra turn does not show the grant")
	}
	if len(g.ExtraTurns) != 0 {
		t.Errorf("the queue still holds %d turns", len(g.ExtraTurns))
	}

	passTurn(t, g)
	if len(g.ScopedEffects) != 0 {
		t.Errorf("%d records survive the extra turn's cleanup", len(g.ScopedEffects))
	}
}

// TestDuringExtraTurnRuleEffectRefusesWhatItCannotBind: no queued turn,
// no scope, no mods.
func TestDuringExtraTurnRuleEffectRefusesWhatItCannotBind(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	refs := takeExtraTurns(t, g, 0, 1)
	mods := []Mod{{Kind: ModDamageCantBePrevented}}
	for name, call := range map[string]func() bool{
		"unqueued ref": func() bool {
			return g.DuringExtraTurnRuleEffectForEffect(refs[0]+5, uuid.Nil, ScopeGame, uuid.Nil, mods, "x")
		},
		"no scope": func() bool {
			return g.DuringExtraTurnRuleEffectForEffect(refs[0], uuid.Nil, ScopeNone, uuid.Nil, mods, "x")
		},
		"unknown scope": func() bool {
			return g.DuringExtraTurnRuleEffectForEffect(refs[0], uuid.Nil, AffectedScope("nowhere"), uuid.Nil, mods, "x")
		},
		"no mods": func() bool {
			return g.DuringExtraTurnRuleEffectForEffect(refs[0], uuid.Nil, ScopeGame, uuid.Nil, nil, "x")
		},
	} {
		ok := true
		g.WithWriteLock(func() { ok = call() })
		if ok {
			t.Errorf("%s: accepted", name)
		}
	}
	if len(g.ExtraTurns[0].Effects) != 0 {
		t.Errorf("refused calls queued %d effects", len(g.ExtraTurns[0].Effects))
	}
}

// TestQueuedExtraTurnEffectsRewindAndAreDeepCopied: a clone keeps the
// queue as it was; queueing another effect on the live game, or editing
// a queued one in place, does not reach the clone.
func TestQueuedExtraTurnEffectsRewindAndAreDeepCopied(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	refs := takeExtraTurns(t, g, 0, 1)
	duringTurn(t, g, refs[0])
	skipStepOf(t, g, refs[0], StepUntap)
	before := g.Clone()

	duringTurn(t, g, refs[0])
	g.ExtraTurns[0].Effects[0].Mods[0].Amount = 9
	if n := len(before.ExtraTurns[0].Effects); n != 1 {
		t.Errorf("the clone's queue grew to %d effects", n)
	}
	if before.ExtraTurns[0].Effects[0].Mods[0].Amount != 0 {
		t.Error("editing a queued effect reached the clone")
	}

	passTurn(t, g)
	if countSkipStepRecords(g) != 0 || len(g.ScopedEffects) == 0 {
		t.Fatalf("setup: skip spent and effects registered, got %d records", len(g.ScopedEffects))
	}
	g.RestoreFrom(before)
	if countSkipStepRecords(g) != 1 {
		t.Errorf("undo: %d skip records, want the 1 the snapshot held", countSkipStepRecords(g))
	}
	if len(g.ExtraTurns) != 1 || len(g.ExtraTurns[0].Effects) != 1 {
		t.Fatalf("undo: queue %+v, want one turn with one effect", g.ExtraTurns)
	}
	// And the redo spends the skip again.
	rock := pushTappedRockForSkipTest(g, 0)
	passTurn(t, g)
	if !rockTappedNow(g, rock) || countSkipStepRecords(g) != 0 {
		t.Errorf("redo: tapped=%v records=%d, want the untap skipped and the record spent",
			rockTappedNow(g, rock), countSkipStepRecords(g))
	}
}

// TestAStepSkipAndQueuedEffectsSurviveASnapshot: both are plain data,
// so a restore point mid-queue carries them.
func TestAStepSkipAndQueuedEffectsSurviveASnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	refs := takeExtraTurns(t, g, 0, 1)
	skipStepOf(t, g, refs[0], StepUntap)
	duringTurn(t, g, refs[0])

	_, restored := roundTrip(t, g)
	if countSkipStepRecords(restored) != 1 {
		t.Fatalf("restored game has %d skip records, want 1", countSkipStepRecords(restored))
	}
	if len(restored.ExtraTurns) != 1 || len(restored.ExtraTurns[0].Effects) != 1 {
		t.Fatalf("restored queue %+v, want one turn with one effect", restored.ExtraTurns)
	}
	rock := pushTappedRockForSkipTest(restored, 0)
	passTurn(t, restored)
	if !rockTappedNow(restored, rock) {
		t.Error("restored game untapped in the skipped untap step")
	}
	if len(restored.DamageCantBePreventedThisTurnLabels()) != 1 {
		t.Error("restored game did not start the queued effect with the turn")
	}
}

// TestASnapshotNamingSkipExtraTurnStepIsRefusedWithoutIt is ADR 0188
// Decision 11: an older binary refuses the file instead of untapping in
// a step the card skipped. The same goes for an unknown mod kind inside a
// queued turn's effects.
func TestASnapshotNamingSkipExtraTurnStepIsRefusedWithoutIt(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	refs := takeExtraTurns(t, g, 0, 1)
	skipStepOf(t, g, refs[0], StepUntap)
	snap := g.CaptureSnapshot()

	spec := modKinds[ModSkipExtraTurnStep]
	delete(modKinds, ModSkipExtraTurnStep)
	defer func() { modKinds[ModSkipExtraTurnStep] = spec }()
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("Restore without skipExtraTurnStep = %v, want ErrUnknownEffectKey", err)
	}
}

func TestAQueuedEffectNamingAnUnknownKindIsRefused(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	refs := takeExtraTurns(t, g, 0, 1)
	duringTurn(t, g, refs[0])
	snap := g.CaptureSnapshot()

	spec := modKinds[ModDamageCantBePrevented]
	delete(modKinds, ModDamageCantBePrevented)
	defer func() { modKinds[ModDamageCantBePrevented] = spec }()
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("Restore with an unknown kind in a queued effect = %v, want ErrUnknownEffectKey", err)
	}
}

// TestAMalformedSkipExtraTurnStepRecordIsRefusedAtRestore: a record with
// no turn is a newer binary's shape.
func TestAMalformedSkipExtraTurnStepRecordIsRefusedAtRestore(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	refs := takeExtraTurns(t, g, 0, 1)
	skipStepOf(t, g, refs[0], StepUntap)
	snap := g.CaptureSnapshot()
	snap.ScopedEffects[0].Mods[0].ExtraTurn = 0
	if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("a skipExtraTurnStep record with no turn = %v, want ErrUnknownEffectKey", err)
	}

	defer func() {
		if recover() == nil {
			t.Error("registering a skipExtraTurnStep mod with no turn did not panic")
		}
	}()
	g.WithWriteLock(func() {
		g.RegisterScopedRuleEffectForEffect(uuid.Nil, ScopeGame, uuid.Nil,
			[]Mod{{Kind: ModSkipExtraTurnStep, Step: StepUntap}}, IndefiniteDuration(), "probe")
	})
}

// TestASkippedUntapStepRunsNoDayNightCheck: night with two spells in the
// previous turn would turn to day at an untap step. The skipped step runs
// no check; the next turn's untap step reads the extra turn.
func TestASkippedUntapStepRunsNoDayNightCheck(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	setDayNight(g, DesignationNight)
	g.WithWriteLock(func() {
		g.SpellsCastThisTurn = map[uuid.UUID]CastTally{g.Seats[0].ID: {Total: 2}}
	})
	refs := takeExtraTurns(t, g, 0, 1)
	skipStepOf(t, g, refs[0], StepUntap)

	passTurn(t, g)
	if !g.IsNight() {
		t.Fatal("the skipped untap step ran the day/night check")
	}
	g.WithWriteLock(func() {
		g.SpellsCastThisTurn = map[uuid.UUID]CastTally{g.Seats[0].ID: {Total: 2}}
	})
	passTurn(t, g)
	if !g.IsDay() {
		t.Error("the next turn's check did not read the extra turn")
	}
}

package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// #6144: "you skip your next untap step" (Avizoa) and "you skip your
// next draw step" (Ivory Gargoyle); skip_next_step.go.
//
// CR 500.11: "To skip a step, phase, or turn is to proceed past it as
// though it didn't exist." CR 614.10: "Once a step, phase, or turn has
// started, it can no longer be skipped—any skip effects will wait until
// the next occurrence." CR 614.10a: "Anything scheduled for the "next"
// occurrence of something waits for the first occurrence that isn't
// skipped. If two effects each cause a player to skip their next
// occurrence, that player must skip the next two".

func skipNextStepOf(t *testing.T, g *Game, player uuid.UUID, step Step) {
	t.Helper()
	var ok bool
	g.WithWriteLock(func() {
		ok = g.SkipNextStepForEffect(player, step, uuid.Nil, "probe")
	})
	if !ok {
		t.Fatal("SkipNextStepForEffect recorded nothing")
	}
}

func nextStepSkipRecords(g *Game) int {
	n := 0
	for _, e := range g.ScopedEffects {
		if _, ok := nextStepSkipRecord(e); ok {
			n++
		}
	}
	return n
}

// turnStepsOf passes to seat's next turn's upkeep and reports whether its
// untap step began, then finishes the turn and reports whether its draw
// step began.
func turnStepsOf(t *testing.T, g *Game, seat int) (untap, draw bool) {
	t.Helper()
	since := eventHighWater(g)
	advanceToStepOfSeat(t, g, seat, StepUpkeep)
	untap = seatStepBegunSince(g, since, g.Seats[seat].ID, StepUntap)
	draw = stepsContain(walkTurnOf(t, g, seat), StepDraw)
	return untap, draw
}

// seatStepBegunSince reports whether `step` of player's turn began after
// event `since`.
func seatStepBegunSince(g *Game, since uint64, player uuid.UUID, step Step) bool {
	for _, ev := range g.Events {
		if ev.Seq > since && ev.Kind == EventStepBegan && ev.Actor == player && ev.Step == step {
			return true
		}
	}
	return false
}

// Avizoa: the next untap step of its controller is skipped, so their
// permanents stay tapped; the one after untaps them.
func TestSkipNextUntapStepKeepsPermanentsTappedOnce(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	me := g.Seats[1].ID
	land := pushTappedPermanent(g, me, "Land", "", "Land", true)
	skipNextStepOf(t, g, me, StepUntap)
	if n := g.NextStepSkipsForLocked(me, StepUntap); n != 1 {
		t.Fatalf("%d untap skips owed, want 1", n)
	}
	untap, draw := turnStepsOf(t, g, 1)
	if untap {
		t.Fatal("the skipped untap step began")
	}
	if !draw {
		t.Error("the turn lost its draw step")
	}
	if !findCard(g, land).Tapped {
		t.Fatal("the land untapped in a skipped untap step")
	}
	if n := nextStepSkipRecords(g); n != 0 {
		t.Fatalf("%d records left, want 0", n)
	}
	if untap, _ := turnStepsOf(t, g, 1); !untap {
		t.Fatal("the untap step after the skipped one did not begin")
	}
	if findCard(g, land).Tapped {
		t.Fatal("the next untap step did not untap the land")
	}
}

// Ivory Gargoyle: exactly the next draw step is skipped; the untap step
// and the draw after are untouched.
func TestSkipNextDrawStepSkipsExactlyOneDraw(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[1].ID
	skipNextStepOf(t, g, me, StepDraw)
	before := len(g.Seats[1].Hand.Cards)
	untap, draw := turnStepsOf(t, g, 1)
	if !untap {
		t.Error("the untap step was skipped too")
	}
	if draw {
		t.Fatal("the skipped draw step began")
	}
	if got := len(g.Seats[1].Hand.Cards); got != before {
		t.Fatalf("hand %d → %d across a skipped draw step", before, got)
	}
	if _, draw := turnStepsOf(t, g, 1); !draw {
		t.Fatal("the draw step after the skipped one did not begin")
	}
}

// Two skips are the next two occurrences (CR 614.10a), one per turn;
// an untap skip leaves the draw step alone and vice versa.
func TestTwoNextUntapSkipsTakeTwoTurnsUntapSteps(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[2].ID
	skipNextStepOf(t, g, me, StepUntap)
	skipNextStepOf(t, g, me, StepUntap)
	for turn := 1; turn <= 2; turn++ {
		if untap, draw := turnStepsOf(t, g, 2); untap || !draw {
			t.Fatalf("turn %d: untap began %v, draw began %v; want false, true", turn, untap, draw)
		}
	}
	if untap, _ := turnStepsOf(t, g, 2); !untap {
		t.Fatal("the third turn's untap step was skipped")
	}
}

// A draw step that has begun can't be skipped (CR 614.10): a skip made
// in the player's own draw step waits for their next one.
func TestANextDrawSkipMadeInThatDrawStepWaits(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceToStepOfSeat(t, g, 1, StepDraw)
	skipNextStepOf(t, g, g.Seats[1].ID, StepDraw)
	advanceToStepOfSeat(t, g, 1, StepPrecombatMain)
	if n := nextStepSkipRecords(g); n != 1 {
		t.Fatalf("%d records after the draw step in progress, want 1", n)
	}
	if _, draw := turnStepsOf(t, g, 1); draw {
		t.Fatal("seat 1's next draw step began")
	}
}

// An untap step another effect skips (Stasis, CR 614.10a) doesn't use
// the record: it waits for the first untap step that happens.
func TestANextUntapSkipWaitsPastAStasisSkippedUntapStep(t *testing.T) {
	const oracle = "test-stasis-probe-6144"
	stubCatalogReplacements(t, map[string][]ReplacementEffect{
		oracle: {{
			Watches: []EventKind{EventStepTransition},
			AppliesTo: func(ev *ReplacementEvent, _ *Game, _ *Card) bool {
				return ev.Kind == RepEventStepTransition && ev.StepTransitionStep == StepUntap
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.Cancel()
				return nil
			},
			PureCancel: true,
			Label:      "probe: players skip their untap steps",
		}},
	})
	g := newActiveGame(t)
	p, q := g.Seats[0], g.Seats[1]
	stasis := pushTappedPermanent(g, q.ID, "Probe", oracle, "Enchantment", false)
	land := pushTappedPermanent(g, p.ID, "Land", "", "Land", true)
	skipNextStepOf(t, g, p.ID, StepUntap)
	passTurnTo(t, g, 0)
	if n := nextStepSkipRecords(g); n != 1 {
		t.Fatalf("%d records after a Stasis-skipped untap step, want 1", n)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == stasis {
				g.Battlefield.Cards = append(g.Battlefield.Cards[:i], g.Battlefield.Cards[i+1:]...)
				break
			}
		}
	})
	passTurnTo(t, g, 0)
	if !findCard(g, land).Tapped {
		t.Fatal("the first untap step that isn't skipped by Stasis was not skipped")
	}
	if n := nextStepSkipRecords(g); n != 0 {
		t.Fatalf("%d records left, want 0", n)
	}
	passTurnTo(t, g, 0)
	if findCard(g, land).Tapped {
		t.Fatal("the untap step after that did not untap the land")
	}
}

// A player who has left writes nothing; a waiting record of a player
// who leaves goes as the next turn begins (CR 800.4k).
func TestANextStepSkipForAPlayerWhoHasLeft(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	g.Seats[3].Eliminated = true
	g.WithWriteLock(func() {
		if g.SkipNextStepForEffect(g.Seats[3].ID, StepDraw, uuid.Nil, "probe") {
			t.Error("a skip was recorded for a player who has left")
		}
	})
	skipNextStepOf(t, g, g.Seats[2].ID, StepUntap)
	g.Seats[2].Eliminated = true
	if n := g.NextStepSkipsForLocked(g.Seats[2].ID, StepUntap); n != 0 {
		t.Errorf("a player who has left owes %d skips", n)
	}
	passTurnTo(t, g, 1)
	if n := nextStepSkipRecords(g); n != 0 {
		t.Errorf("%d records left for a player who has left, want 0", n)
	}
}

// The record is plain data: a restore point keeps the pending skip, and
// an older binary refuses the file rather than give the step back.
func TestNextStepSkipsSurviveASnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	skipNextStepOf(t, g, g.Seats[1].ID, StepDraw)
	_, restored := roundTrip(t, g)
	if n := restored.NextStepSkipsForLocked(restored.Seats[1].ID, StepDraw); n != 1 {
		t.Fatalf("restored game owes %d draw skips, want 1", n)
	}
	if _, draw := turnStepsOf(t, restored, 1); draw {
		t.Fatal("the restored skip did not hold")
	}

	snap := g.CaptureSnapshot()
	spec, ok := modKinds[ModSkipNextStep]
	if !ok {
		t.Fatal("skipNextStep is not in the mod-kind table")
	}
	delete(modKinds, ModSkipNextStep)
	defer func() { modKinds[ModSkipNextStep] = spec }()
	if _, err := snap.Restore(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("Restore without skipNextStep = %v, want ErrUnknownEffectKey", err)
	}
}

func TestASkipNextStepModProblem(t *testing.T) {
	for _, m := range []Mod{
		{Kind: ModSkipNextStep, Step: StepUntap},
		{Kind: ModSkipNextStep, Player: uuid.New(), Step: StepUpkeep},
		{Kind: ModSkipNextStep, Player: uuid.New(), Step: StepDraw, Amount: 2},
	} {
		if skipNextStepModProblem(m) == "" {
			t.Errorf("%+v accepted", m)
		}
	}
	if p := skipNextStepModProblem(Mod{Kind: ModSkipNextStep, Player: uuid.New(), Step: StepUntap}); p != "" {
		t.Errorf("a sound mod refused: %s", p)
	}
	g := newActiveGame(t)
	skipNextStepOf(t, g, g.Seats[1].ID, StepUntap)
	snap := g.CaptureSnapshot()
	snap.ScopedEffects[0].Mods[0].Step = StepEnd
	if _, err := snap.RestoreStrict(); !errors.Is(err, ErrUnknownEffectKey) {
		t.Errorf("a restore point naming the end step = %v, want ErrUnknownEffectKey", err)
	}
}

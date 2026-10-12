package game

import (
	"testing"

	"github.com/google/uuid"
)

func TestDrawStepTallySurvivesCloneUndoAndDiskRestore(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	advanceTo(t, g, StepDraw)
	p := g.Seats[0]
	before := g.DrawnThisStep[p.ID]
	if before != 1 {
		t.Fatalf("initial draw-step tally %d, want 1", before)
	}
	clone := g.Clone()
	g.WithWriteLock(func() {
		if err := g.DrawNForEffect(p.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	if clone.DrawnThisStep[p.ID] != before {
		t.Fatal("a live draw changed the clone's tally")
	}
	g.RestoreFrom(clone)
	if g.DrawnThisStep[p.ID] != before {
		t.Fatal("undo did not restore the step tally")
	}
	snapshot := g.CaptureSnapshot()
	restored, err := snapshot.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	if restored.DrawnThisStep[p.ID] != before {
		t.Fatal("disk restore lost the step tally")
	}
	restored.DrawnThisStep[p.ID]++
	if snapshot.DrawnThisStep[p.ID] != before {
		t.Fatal("restored tally shares the snapshot's map")
	}
}

func TestLegacyDrawStepTallyReconstructsOnlyCurrentStep(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	g.Turn.Step = StepDraw
	g.DrawnThisStep = nil
	g.Events = []Event{
		{Kind: EventStepBegan, Step: StepDraw},
		{Kind: EventDrawCard, Actor: p.ID},
		{Kind: EventStepBegan, Step: StepDraw},
		{Kind: EventDrawCard, Actor: p.ID},
		{Kind: EventDrawCard, Actor: p.ID},
	}
	g.restoreDrawStepCountLocked()
	if got := g.DrawnThisStep[p.ID]; got != 2 {
		t.Fatalf("restored tally %d, want two draws in the latest draw step", got)
	}
}

func TestFailedExpandedDrawStillFinishesSiblingAndThen(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p, other := g.Seats[0], g.Seats[1]
	stockLibrary(p, 0)
	stockLibrary(other, 5)
	before := other.Hand.Size()
	then := RegisterDrawThen("test-failed-expanded-draw-then", func(g *Game, d DrawThen) error {
		g.playerByIDLocked(d.Player).Life++
		return nil
	})
	life := p.Life
	g.WithWriteLock(func() {
		g.RegisterReplacementForTest(drawCountReplacement(p.ID, timesTwoForTest, "double"))
		g.RegisterReplacementForTest(ReplacementEffect{
			Watches: []EventKind{EventDrawCard},
			AppliesTo: func(ev *ReplacementEvent, g *Game, _ *Card) bool {
				return ev.DrawPlayer == p.ID && g.playerByIDLocked(p.ID).AttemptedEmptyDraw
			},
			Replace: func(ev *ReplacementEvent, _ *Game, _ *Card) error {
				ev.DrawPlayer = other.ID
				return nil
			},
			Controller: func(*ReplacementEvent, *Game, *Card) uuid.UUID { return p.ID },
		})
		if err := g.DrawNThenForEffect(p.ID, 1, DrawThen{Ref: then, Player: p.ID}); err != nil {
			t.Fatal(err)
		}
	})
	if other.Hand.Size() != before+1 || p.Life != life+1 {
		t.Fatal("failed first draw skipped its redirected sibling or repeated its then")
	}
}

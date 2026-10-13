package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func countCrimeEventsForTest(events []game.Event) int {
	n := 0
	for _, ev := range events {
		if ev.Kind == game.EventBecomesTarget && ev.CrimeCommitted {
			n++
		}
	}
	return n
}

func TestCopyingABoltAfterMagdaEnteredDoesNotCommitAnotherCrime(t *testing.T) {
	g := newCatalogGame(t)
	me, opponent := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetPlayer(opponent.ID))
	if got := countCrimeEventsForTest(g.Events); got != 1 {
		t.Fatalf("casting Lightning Bolt at an opponent emitted %d crimes, want one", got)
	}

	// Enter after the Bolt's announcement so a second crime event caused
	// by copying it cannot be hidden by Magda's once-per-turn limit.
	magda := b12Push(g, me.ID, "Magda, the Hoardmaster", "Legendary Creature — Dwarf Berserker", b18MagdaTheHoardmasterOrcl, 2, 2)
	magdaTriggers := func() int {
		n := triggersOnStackFrom(g, magda)
		for _, item := range g.PendingTriggers {
			if item != nil && item.SourceCardID == magda {
				n++
			}
		}
		return n
	}
	if got := magdaTriggers(); got != 0 {
		t.Fatalf("Magda has %d crime triggers before copying, want none", got)
	}

	becomesBefore := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventBecomesTarget {
			becomesBefore++
		}
	}
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(bolt, me.ID, false, nil); err != nil {
			t.Errorf("CopySpellForEffect: %v", err)
		}
	})
	if got := countCrimeEventsForTest(g.Events); got != 1 {
		t.Errorf("copy emitted another crime event; total %d, want one", got)
	}
	if got := magdaTriggers(); got != 0 {
		t.Errorf("copy caused Magda to have %d crime triggers, want none", got)
	}
	becomesAfter := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventBecomesTarget {
			becomesAfter++
		}
	}
	if becomesAfter <= becomesBefore {
		t.Errorf("the copy emitted no ordinary becomes-target event (before %d, after %d)", becomesBefore, becomesAfter)
	}
}

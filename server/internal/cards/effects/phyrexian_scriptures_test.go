package effects

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Scriptures (#2155): chapter I's artifact grant has no
// duration, so the creature it names survives chapter II, which spares
// artifacts, while a bystander does not.
func TestPhyrexianScripturesChapterOneMakesItAnArtifactChapterTwoSparesIt(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat].ID
	saved := b12Creature(g, me, "Saved Bear", "Creature — Bear", 2, 2)
	doomed := b12Creature(g, me, "Doomed Bear", "Creature — Bear", 2, 2)
	saga := castCatalogSpell(t, g, "Phyrexian Scriptures", "Enchantment — Saga",
		"11173ad3-c007-478f-bce0-d756eac07ccb", nil)
	passPriorityAroundTable(t, g)
	answerPickTarget(t, g, saved)
	passPriorityAroundTable(t, g)
	if !slices.Contains(effectiveTypes(t, g, saved), "Artifact") {
		t.Fatalf("chapter I: types %v, want Artifact added", effectiveTypes(t, g, saved))
	}
	if n := counterCount(g, saved, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", n)
	}
	advanceToPrecombatMainOf(t, g, seat)
	passPriorityAroundTable(t, g)
	if l := loreCountersOn(g, saga); l != 2 {
		t.Fatalf("lore = %d, want chapter II reached", l)
	}
	if !g.Battlefield.Contains(saved) {
		t.Error("the artifact-ified creature should survive chapter II")
	}
	if g.Battlefield.Contains(doomed) {
		t.Error("a nonartifact creature should die to chapter II")
	}
}

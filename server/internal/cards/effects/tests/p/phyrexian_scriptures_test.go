package testsp

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Scriptures (#2155): chapter I's artifact grant has no
// duration, so the creature it names survives chapter II, which spares
// artifacts, while a bystander does not.
func TestPhyrexianScripturesChapterOneMakesItAnArtifactChapterTwoSparesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat].ID
	saved := effectstest.B12Creature(g, me, "Saved Bear", "Creature — Bear", 2, 2)
	doomed := effectstest.B12Creature(g, me, "Doomed Bear", "Creature — Bear", 2, 2)
	saga := effectstest.CastCatalogSpell(t, g, "Phyrexian Scriptures", "Enchantment — Saga",
		"11173ad3-c007-478f-bce0-d756eac07ccb", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, saved)
	effectstest.PassPriorityAroundTable(t, g)
	if !slices.Contains(effectstest.EffectiveTypes(t, g, saved), "Artifact") {
		t.Fatalf("chapter I: types %v, want Artifact added", effectstest.EffectiveTypes(t, g, saved))
	}
	if n := effectstest.CounterCount(g, saved, game.CounterPlusOne); n != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", n)
	}
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if l := effectstest.LoreCountersOn(g, saga); l != 2 {
		t.Fatalf("lore = %d, want chapter II reached", l)
	}
	if !g.Battlefield.Contains(saved) {
		t.Error("the artifact-ified creature should survive chapter II")
	}
	if g.Battlefield.Contains(doomed) {
		t.Error("a nonartifact creature should die to chapter II")
	}
}

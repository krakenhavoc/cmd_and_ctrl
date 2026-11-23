package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// activateHallOfEchoes pays {5} and activates Hall's copy ability on
// `target`, leaving it on the stack.
func activateHallOfEchoes(t *testing.T, g *game.Game, me uuid.UUID, hall, target uuid.UUID) {
	t.Helper()
	effectstest.DcMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me, hall, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("activate Hall of Echoes: %v", err)
	}
}

func legendRuleRecords(g *game.Game) int {
	n := 0
	g.WithWriteLock(func() {
		for i := range g.ScopedEffects {
			for _, m := range g.ScopedEffects[i].Mods {
				if m.Kind == game.ModLegendRuleDoesntApply {
					n++
				}
			}
		}
	})
	return n
}

// TestHallOfEchoesBecomesACopyAndSuspendsTheLegendRuleForYouOnly is the
// card end to end: the land becomes a second copy of your legend, the
// pair stays, and an opponent's pair does not.
func TestHallOfEchoesBecomesACopyAndSuspendsTheLegendRuleForYouOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hall := effectstest.DcSeed(g, me.ID, "Hall of Echoes", "Land", effectstest.OracleHallOfEchoes, 0, 0)
	legend := effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)

	activateHallOfEchoes(t, g, me.ID, hall, legend)
	effectstest.PassPriorityAroundTable(t, g)
	// The opponent's pair arrives after the ability resolved: an open
	// legend-rule prompt would stop priority from passing at all.
	c := effectstest.B12Creature(g, opp.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	d := effectstest.B12Creature(g, opp.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)
	g.RunStateChecksForTest()

	if got := effectstest.DcCard(t, g, hall); got.Name != effectstest.LegendName || !got.IsCreature() || got.IsLand() {
		t.Fatalf("Hall of Echoes is %q (creature %v, land %v), want a copy of %s", got.Name, got.IsCreature(), got.IsLand(), effectstest.LegendName)
	}
	if !effectstest.OnBoard(g, hall) || !effectstest.OnBoard(g, legend) || effectstest.LegendPrompt(g, me.ID) != nil {
		t.Error(effectstest.BothLegendsMustStay + " while the rule is suspended")
	}
	if effectstest.LegendPrompt(g, opp.ID) == nil || !effectstest.OnBoard(g, c) || !effectstest.OnBoard(g, d) {
		t.Error("the opponent's pair was exempted by Hall of Echoes; the exemption is for permanents you control")
	}
}

// TestHallOfEchoesExemptionEndsAtCleanup: the rule is back the moment
// the turn ends, for a pair Hall had nothing to do with.
func TestHallOfEchoesExemptionEndsAtCleanup(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hall := effectstest.DcSeed(g, me.ID, "Hall of Echoes", "Land", effectstest.OracleHallOfEchoes, 0, 0)
	legend := effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)

	activateHallOfEchoes(t, g, me.ID, hall, legend)
	effectstest.PassPriorityAroundTable(t, g)
	// A pair that has nothing to do with Hall, arriving while the
	// exemption stands (a prompt open before the ability resolved
	// would have stopped priority passing).
	a := effectstest.B12Creature(g, me.ID, "Other Legend", effectstest.LegendTypeLine, 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Other Legend", effectstest.LegendTypeLine, 1, 1)
	g.RunStateChecksForTest()
	if effectstest.LegendPrompt(g, me.ID) != nil || !effectstest.OnBoard(g, a) || !effectstest.OnBoard(g, b) {
		t.Fatal("the unrelated pair was not exempted this turn")
	}
	if n := legendRuleRecords(g); n != 1 {
		t.Fatalf("%d exemption records on the table, want 1", n)
	}

	effectstest.DcCleanup(g)
	if n := legendRuleRecords(g); n != 0 {
		t.Fatalf("%d exemption records survived cleanup", n)
	}
	if got := effectstest.DcCard(t, g, hall); got.Name != "Hall of Echoes" || !got.IsLand() {
		t.Errorf("Hall is %q after cleanup, want the land back", got.Name)
	}
	g.RunStateChecksForTest()
	if effectstest.LegendPrompt(g, me.ID) == nil {
		t.Error("the legend rule did not apply again after cleanup")
	}
}

// TestHallOfEchoesFizzlesWithNoLegalTarget: the ability's only target
// is gone on resolution, so neither sentence happens (CR 608.2b).
func TestHallOfEchoesFizzlesWithNoLegalTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hall := effectstest.DcSeed(g, me.ID, "Hall of Echoes", "Land", effectstest.OracleHallOfEchoes, 0, 0)
	legend := effectstest.B12Creature(g, me.ID, effectstest.LegendName, effectstest.LegendTypeLine, 2, 2)

	activateHallOfEchoes(t, g, me.ID, hall, legend)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(legend); err != nil {
			t.Fatalf("destroy the target: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	a := effectstest.B12Creature(g, me.ID, "Other Legend", effectstest.LegendTypeLine, 1, 1)
	effectstest.B12Creature(g, me.ID, "Other Legend", effectstest.LegendTypeLine, 1, 1)
	g.RunStateChecksForTest()

	if n := legendRuleRecords(g); n != 0 {
		t.Errorf("%d exemption records written by an ability with no legal target", n)
	}
	if got := effectstest.DcCard(t, g, hall); got.Name != "Hall of Echoes" {
		t.Errorf("Hall is %q, want it unchanged", got.Name)
	}
	if effectstest.LegendPrompt(g, me.ID) == nil || !effectstest.OnBoard(g, a) {
		t.Error("the legend rule should have applied to the other pair")
	}
}

// TestHallOfEchoesIsAManaLandUntilItCopiesSomething pins the printed
// first line next to the second.
func TestHallOfEchoesIsAManaLandUntilItCopiesSomething(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.OracleHallOfEchoes)
	if !ok {
		t.Fatal("Hall of Echoes is not registered")
	}
	if spec.Completeness != effects.CompletenessFull {
		t.Errorf("Completeness = %v, want Full", spec.Completeness)
	}
	if len(spec.ManaAbilities) != 1 || len(spec.Activated) != 1 {
		t.Fatalf("%d mana abilities, %d activated; want 1 and 1", len(spec.ManaAbilities), len(spec.Activated))
	}
}

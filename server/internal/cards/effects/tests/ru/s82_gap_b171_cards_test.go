package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 GAP batch 171 (Jumpstart 2022 instants, #5948).

const (
	gapB171HowlOracle   = "a1180796-6e26-4b45-9263-130620d702e5"
	gapB171GrowthOracle = "acd81d2c-e438-44db-86eb-32fc9cb3fd8f"
	gapB171DevilsOracle = "caeb508a-c1fd-45b5-a14e-e0ac7321c9a9"
	gapB171WorthyOracle = "008c92bc-3022-4d34-aa04-584576ff8f8e"
	gapB171OutnumOracle = "b72efd67-a234-44af-8296-3e182292c7e6"
)

func TestGapB171CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB171HowlOracle: "Predator's Howl", gapB171GrowthOracle: "Rampaging Growth",
		gapB171DevilsOracle: "Dance with Devils", gapB171WorthyOracle: "Deem Worthy",
		gapB171OutnumOracle: "Outnumber",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, name %q, completeness %v, caveats %v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestGapB171PredatorsHowlMakesOneWolfOrThreeWithMorbid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Predator's Howl", "Instant", gapB171HowlOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B28TokensNamed(g, me.ID, "Wolf"); n != 1 {
		t.Fatalf("without morbid: %d Wolves, want 1", n)
	}
	doomed := effectstest.B12Creature(g, g.Seats[1].ID, "Doomed", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(doomed, doomed, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.CastCatalogSpell(t, g, "Predator's Howl", "Instant", gapB171HowlOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B28TokensNamed(g, me.ID, "Wolf"); n != 4 {
		t.Errorf("with morbid: %d Wolves in all, want 4 (1 + 3)", n)
	}
}

func TestGapB171RampagingGrowthFetchesAndAnimatesTheLandUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.PushLibraryCard(me, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Rampaging Growth", "Instant", gapB171GrowthOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if open := effectstest.ChkB1OpenSearch(g); open != nil {
		if err := g.ResolveSearchLibrary(open.ID, open.Chooser, []uuid.UUID{forest}); err != nil {
			t.Fatalf("ResolveSearchLibrary: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	c, ok := effectstest.BattlefieldCard(g, forest)
	if !ok {
		t.Fatalf("the Forest is in %v, want the battlefield", effectstest.ZoneOf(g, forest))
	}
	if c.Tapped {
		t.Error("the land entered tapped")
	}
	if p, tough := effectstest.PtOf(t, g, forest); p != 4 || tough != 3 {
		t.Errorf("the land is %d/%d, want 4/3", p, tough)
	}
	eff := effectstest.B12Card(t, g, forest)
	if !eff.IsCreature() || !eff.IsLand() || !eff.HasSubtype("Insect") {
		t.Errorf("the land is creature %v land %v Insect %v, want all", eff.IsCreature(), eff.IsLand(), eff.HasSubtype("Insect"))
	}
	for _, kw := range []string{"reach", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, forest, kw) {
			t.Errorf("the animated land does not have %s", kw)
		}
	}
}

func TestGapB171RampagingGrowthWithNoBasicLandAnimatesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	before := len(g.Battlefield.Cards)
	effectstest.CastCatalogSpell(t, g, "Rampaging Growth", "Instant", gapB171GrowthOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	_ = me
	if got := len(g.Battlefield.Cards); got != before {
		t.Errorf("the battlefield went from %d to %d permanents", before, got)
	}
}

func TestGapB171DanceWithDevilsMakesTwoDevilsThatPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Dance with Devils", "Instant", gapB171DevilsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	devils := []uuid.UUID{}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Devil" && effects.IsToken(c) {
			devils = append(devils, c.InstanceID)
		}
	}
	if len(devils) != 2 {
		t.Fatalf("%d Devils, want 2", len(devils))
	}
	killer := effectstest.B12Creature(g, opp.ID, "Killer", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(killer, devils[0], 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if pick := effectstest.LatestPickTarget(g, me.ID); pick == nil {
		t.Fatal("the dying Devil's trigger asked for no target")
	}
}

func TestGapB171DeemWorthyDealsSevenAndCyclesForTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 7, 7)
	effectstest.CastCatalogSpell(t, g, "Deem Worthy", "Instant", gapB171WorthyOracle, effectstest.CardTarget(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if effectstest.BattlefieldHas(g, ogre) {
		t.Error("a 7/7 survived 7 damage")
	}
}

func TestGapB171DeemWorthyCycleTriggerDealsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	imp := effectstest.B12Creature(g, opp.ID, "Imp", "Creature — Imp", 2, 2)
	effectstest.CycleFromHand(t, g, "Deem Worthy", "Instant", gapB171WorthyOracle, "{C}{C}{C}{R}")
	effectstest.PickTriggerTarget(t, g, me.ID, imp)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if effectstest.BattlefieldHas(g, imp) {
		t.Error("the cycled Deem Worthy did not kill a 2/2")
	}
}

func TestGapB171OutnumberCountsYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	}
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	victim := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 1, 3)
	effectstest.CastCatalogSpell(t, g, "Outnumber", "Instant", gapB171OutnumOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if effectstest.BattlefieldHas(g, victim) {
		t.Error("3 damage did not kill a 1/3 (three creatures are yours)")
	}
}

func TestGapB171OutnumberDoesNotKillWithTooFewCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 1, 3)
	effectstest.CastCatalogSpell(t, g, "Outnumber", "Instant", gapB171OutnumOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if !effectstest.BattlefieldHas(g, victim) {
		t.Error("1 damage killed a 1/3")
	}
}

package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b49_test.go — Star Trek Commander batch 49 (#4974): Mutavault,
// Shattered Landscape, Tranquil Landscape and Run Amok.

const (
	trekB49MutavaultOracle = "6b3cc59a-7ea5-4eb5-9bf9-5a9c07f80e2b"
	trekB49ShatteredOracle = "7fbad3f2-66f6-4e3a-b9e0-1ddbcd94d42a"
	trekB49TranquilOracle  = "d4eb65d5-99fd-4daf-b7d3-8ebf99ee9c61"
	trekB49RunAmokOracle   = "d463de75-8d3c-4077-a2cc-c304a63acc69"
)

func TestTrekB49CardsAreFull(t *testing.T) {
	want := map[string]string{
		trekB49MutavaultOracle: "Mutavault",
		trekB49ShatteredOracle: "Shattered Landscape",
		trekB49TranquilOracle:  "Tranquil Landscape",
		trekB49RunAmokOracle:   "Run Amok",
	}
	for oracle, name := range want {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestTrekB49MutavaultBecomesATwoTwoWithAllCreatureTypes(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	vault := effectstest.PushCatalogPermanent(g, me.ID, "Mutavault", "Land", trekB49MutavaultOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{1} paid from an empty pool")
	}
	effectstest.ApaMana(me, "C")
	effectstest.B16Activate(t, g, me.ID, vault, 0, game.ActivateAbilityParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.EffectiveCardForTest(t, g, vault)
	if !c.IsCreature() || !c.IsLand() {
		t.Errorf("after animating: creature %v, land %v, want both", c.IsCreature(), c.IsLand())
	}
	for _, sub := range []string{"Elf", "Human", "Shapeshifter", "Goblin"} {
		if !c.HasSubtype(sub) {
			t.Errorf("the animated land is not a %s", sub)
		}
	}
	if c.HasSubtype("Forest") {
		t.Error("all creature types gave it a land type")
	}
	if p, tough := effectstest.EffectivePower(t, g, vault), effectstest.EffectiveToughness(t, g, vault); p != 2 || tough != 2 {
		t.Errorf("P/T %d/%d, want 2/2", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, vault, "changeling") {
		t.Error("the animated land gained changeling (ruling 2008-04-01)")
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	if c := effectstest.EffectiveCardForTest(t, g, vault); c.IsCreature() || c.HasSubtype("Elf") || !c.IsLand() {
		t.Errorf("after cleanup: creature %v, Elf %v, land %v, want a plain land", c.IsCreature(), c.HasSubtype("Elf"), c.IsLand())
	}
}

func TestTrekB49LandscapesFetchTheirThreeBasicsTapped(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		fetch        []string
		refuse       string
	}{
		{"Shattered Landscape", trekB49ShatteredOracle, []string{"Mountain", "Plains", "Swamp"}, "Island"},
		{"Tranquil Landscape", trekB49TranquilOracle, []string{"Forest", "Plains", "Island"}, "Swamp"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			land := effectstest.PushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			for _, sub := range append(append([]string{}, tc.fetch...), tc.refuse) {
				effectstest.StapleLibraryCard(me, "Basic "+sub, "Basic Land — "+sub)
			}
			if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("fetch: %v", err)
			}
			if _, still := effectstest.BattlefieldCard(g, land); still {
				t.Error("the land is sacrificed as a cost")
			}
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt")
			}
			if effectstest.SearchOptionNamed(g, c, "Basic "+tc.refuse) != uuid.Nil {
				t.Errorf("the search offered a basic %s", tc.refuse)
			}
			for _, sub := range tc.fetch {
				if effectstest.SearchOptionNamed(g, c, "Basic "+sub) == uuid.Nil {
					t.Errorf("the search did not offer a basic %s", sub)
				}
			}
			effectstest.AnswerSearchNamed(t, g, me.ID, "Basic "+tc.fetch[0])
			effectstest.PassPriorityAroundTable(t, g)
			found := false
			for _, bc := range g.Battlefield.Cards {
				if bc.Name == "Basic "+tc.fetch[0] {
					found = true
					if !bc.Tapped {
						t.Error("the fetched land should enter tapped")
					}
				}
			}
			if !found {
				t.Error("the fetched land did not reach the battlefield")
			}
		})
	}
}

func TestTrekB49LandscapesCycleForTheirCost(t *testing.T) {
	for _, tc := range []struct{ name, oracle, wrong, right string }{
		{"Shattered Landscape", trekB49ShatteredOracle, "{G}{W}{U}", "{R}{W}{B}"},
		{"Tranquil Landscape", trekB49TranquilOracle, "{R}{W}{B}", "{G}{W}{U}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
			id := effectstest.PushCatalogHandCard(me, tc.name, "Land", tc.oracle)
			if err := g.AddManaForEffect(me.ID, uuid.Nil, tc.wrong); err != nil {
				t.Fatalf("AddManaForEffect: %v", err)
			}
			if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatalf("cycled for %s, the cost is %s", tc.wrong, tc.right)
			}
			if err := g.AddManaForEffect(me.ID, uuid.Nil, tc.right); err != nil {
				t.Fatalf("AddManaForEffect: %v", err)
			}
			handBefore := me.Hand.Size()
			if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
				t.Fatalf("cycle: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Graveyard.Contains(id) {
				t.Error("the cycled land should be in the graveyard")
			}
			if me.Hand.Size() != handBefore {
				t.Errorf("hand %d -> %d, want the cycled card replaced by a draw", handBefore, me.Hand.Size())
			}
		})
	}
}

func TestTrekB49RunAmokPumpsAnAttackerWithTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.MscM1Attacking(g, bear, opp.ID)
	if err := effectstest.CastCatalogSpellErr(t, g, "Run Amok", "Instant", trekB49RunAmokOracle, effectstest.B16TargetCard(idle)); err == nil {
		t.Error("a creature that is not attacking was accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Run Amok", "Instant", trekB49RunAmokOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tg := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 5 || tg != 5 {
		t.Errorf("attacker is %d/%d, want 5/5", p, tg)
	}
	effectstest.AssertKeywords(t, g, bear, "trample")
	if p := effectstest.EffectivePower(t, g, idle); p != 2 {
		t.Errorf("the idle bear is pumped to power %d", p)
	}
}

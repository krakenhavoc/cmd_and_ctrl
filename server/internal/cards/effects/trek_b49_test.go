package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestTrekB49MutavaultBecomesATwoTwoWithAllCreatureTypes(t *testing.T) {
	g, me, _ := spendTable(t)
	vault := pushCatalogPermanent(g, me.ID, "Mutavault", "Land", trekB49MutavaultOracle, false)
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("{1} paid from an empty pool")
	}
	apaMana(me, "C")
	b16Activate(t, g, me.ID, vault, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	c := effectiveCardForTest(t, g, vault)
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
	if p, tough := effectivePower(t, g, vault), effectiveToughness(t, g, vault); p != 2 || tough != 2 {
		t.Errorf("P/T %d/%d, want 2/2", p, tough)
	}
	if hasEffectiveKeyword(t, g, vault, "changeling") {
		t.Error("the animated land gained changeling (ruling 2008-04-01)")
	}
	advancePastCleanupOf(t, g, 0)
	if c := effectiveCardForTest(t, g, vault); c.IsCreature() || c.HasSubtype("Elf") || !c.IsLand() {
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
			g := newCatalogGame(t)
			me := g.Seats[0]
			land := pushCatalogPermanent(g, me.ID, tc.name, "Land", tc.oracle, false)
			for _, sub := range append(append([]string{}, tc.fetch...), tc.refuse) {
				stapleLibraryCard(me, "Basic "+sub, "Basic Land — "+sub)
			}
			if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("fetch: %v", err)
			}
			if _, still := battlefieldCard(g, land); still {
				t.Error("the land is sacrificed as a cost")
			}
			passPriorityAroundTable(t, g)
			c := searchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt")
			}
			if searchOptionNamed(g, c, "Basic "+tc.refuse) != uuid.Nil {
				t.Errorf("the search offered a basic %s", tc.refuse)
			}
			for _, sub := range tc.fetch {
				if searchOptionNamed(g, c, "Basic "+sub) == uuid.Nil {
					t.Errorf("the search did not offer a basic %s", sub)
				}
			}
			answerSearchNamed(t, g, me.ID, "Basic "+tc.fetch[0])
			passPriorityAroundTable(t, g)
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
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceTo(t, g, game.StepPrecombatMain)
			id := pushCatalogHandCard(me, tc.name, "Land", tc.oracle)
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
			passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	mscM1Attacking(g, bear, opp.ID)
	if err := castCatalogSpellErr(t, g, "Run Amok", "Instant", trekB49RunAmokOracle, b16TargetCard(idle)); err == nil {
		t.Error("a creature that is not attacking was accepted")
	}
	castCatalogSpell(t, g, "Run Amok", "Instant", trekB49RunAmokOracle, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tg != 5 {
		t.Errorf("attacker is %d/%d, want 5/5", p, tg)
	}
	assertKeywords(t, g, bear, "trample")
	if p := effectivePower(t, g, idle); p != 2 {
		t.Errorf("the idle bear is pumped to power %d", p)
	}
}

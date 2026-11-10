package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const firdochCoreOracle = "126c75c0-414b-4e55-8747-2b5566ce716f"

func TestFirdochCoreIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(firdochCoreOracle)
	if !ok || spec.Name != "Firdoch Core" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// A kindred artifact with changeling is every creature type without
// being a creature; it taps for any colour.
func TestFirdochCoreIsEveryTypeAndTapsForAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	core := b12Push(g, me.ID, "Firdoch Core", "Kindred Artifact — Shapeshifter", firdochCoreOracle, 0, 0)
	c := ecl5Card(t, g, core)
	if c.IsCreature() {
		t.Fatal("the Core is a creature before it is animated")
	}
	for _, sub := range []string{"Elf", "Goblin", "Treefolk", "Shapeshifter"} {
		if !c.HasSubtype(sub) {
			t.Errorf("changeling: the Core is not a %s", sub)
		}
	}
	if err := g.ActivateManaAbility(me.ID, core, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	pick := riderLatestManaPick(g, me.ID)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("pick = %+v, want all five colours", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if me.ManaPool.Total() != 1 || me.ManaPool.Units()[0].Color != "U" {
		t.Errorf("pool = %+v, want one {U}", me.ManaPool)
	}
}

// {4}: a 4/4 artifact creature until end of turn, still every type. Three
// mana can't pay for it.
func TestFirdochCoreBecomesAFourFourUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	core := b12Push(g, me.ID, "Firdoch Core", "Kindred Artifact — Shapeshifter", firdochCoreOracle, 0, 0)
	advanceToMain(t, g)

	b06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, core, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("three mana paid for {4}")
	}
	b06AddMana(me, "C")
	b16Activate(t, g, me.ID, core, 0, game.ActivateAbilityParams{Strict: true})
	c := ecl5Card(t, g, core)
	if !c.IsCreature() || !c.IsArtifact() {
		t.Fatalf("animated Core types = %q, want an artifact creature", c.TypeLine)
	}
	if p, tough := effectivePower(t, g, core), effectiveToughness(t, g, core); p != 4 || tough != 4 {
		t.Errorf("animated Core is %d/%d, want 4/4", p, tough)
	}
	if !c.HasSubtype("Kithkin") {
		t.Error("the animated Core lost its creature types")
	}

	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	aangAdvanceToMain(t, g, 1)
	if ecl5Card(t, g, core).IsCreature() {
		t.Error("the Core is still a creature after the turn ended")
	}
}

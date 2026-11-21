package testsl

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const lavaleaperOracle = "a107fb84-6874-4964-a729-4425dad98330"

func TestLavaleaperIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(lavaleaperOracle)
	if !ok || spec.Name != "Lavaleaper" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// Every creature, an opponent's too, has haste.
func TestLavaleaperGivesEveryCreatureHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Fatal("the Bear has haste before Lavaleaper")
	}
	leaper := effectstest.PushCatalogPermanent(g, me.ID, "Lavaleaper", "Creature — Elemental", lavaleaperOracle, false)
	for _, id := range []uuid.UUID{theirs, leaper} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%v lacks haste", id)
		}
	}
}

// A basic land, anyone's, makes one more of its mana; a nonbasic land
// makes nothing extra.
func TestLavaleaperDoublesOnlyBasicLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Lavaleaper", "Creature — Elemental", lavaleaperOracle, false)
	mine := effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	theirs := effectstest.B12Permanent(g, them.ID, "Island", "Basic Land — Island")
	effectstest.TapForMana(t, g, me.ID, mine)
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("my pool = %v, want two red", got)
	}
	effectstest.TapForMana(t, g, them.ID, theirs)
	if got := effectstest.SortedPool(them); len(got) != 2 || got[0] != "U" || got[1] != "U" {
		t.Errorf("their pool = %v, want two blue", got)
	}

	vault := effectstest.PushCatalogPermanent(g, them.ID, "Mutavault", "Land", effects.MutavaultOracleID, false)
	before := them.ManaPool.Total()
	effectstest.TapForMana(t, g, them.ID, vault)
	if got := them.ManaPool.Total(); got != before+1 {
		t.Errorf("a nonbasic land made %d mana, want 1", got-before)
	}
}

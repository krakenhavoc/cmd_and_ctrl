package effects

import (
	"testing"

	"github.com/google/uuid"
)

const lavaleaperOracle = "a107fb84-6874-4964-a729-4425dad98330"

func TestLavaleaperIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(lavaleaperOracle)
	if !ok || spec.Name != "Lavaleaper" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// Every creature, an opponent's too, has haste.
func TestLavaleaperGivesEveryCreatureHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	if hasEffectiveKeyword(t, g, theirs, "haste") {
		t.Fatal("the Bear has haste before Lavaleaper")
	}
	leaper := pushCatalogPermanent(g, me.ID, "Lavaleaper", "Creature — Elemental", lavaleaperOracle, false)
	for _, id := range []uuid.UUID{theirs, leaper} {
		if !hasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%v lacks haste", id)
		}
	}
}

// A basic land, anyone's, makes one more of its mana; a nonbasic land
// makes nothing extra.
func TestLavaleaperDoublesOnlyBasicLands(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Lavaleaper", "Creature — Elemental", lavaleaperOracle, false)
	mine := b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	theirs := b12Permanent(g, them.ID, "Island", "Basic Land — Island")
	tapForMana(t, g, me.ID, mine)
	if got := sortedPool(me); len(got) != 2 || got[0] != "R" || got[1] != "R" {
		t.Errorf("my pool = %v, want two red", got)
	}
	tapForMana(t, g, them.ID, theirs)
	if got := sortedPool(them); len(got) != 2 || got[0] != "U" || got[1] != "U" {
		t.Errorf("their pool = %v, want two blue", got)
	}

	vault := pushCatalogPermanent(g, them.ID, "Mutavault", "Land", mutavaultOracleID, false)
	before := them.ManaPool.Total()
	tapForMana(t, g, them.ID, vault)
	if got := them.ManaPool.Total(); got != before+1 {
		t.Errorf("a nonbasic land made %d mana, want 1", got-before)
	}
}

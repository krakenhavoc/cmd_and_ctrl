package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// raffine_a_tidebinder_test.go — Tishana's Tidebinder (#3212, slice raffine-a).

const tidebinderOracle = "2993dc7d-723d-4a9b-94bd-4bb02a9f7243"

func TestTidebinderIsRegisteredAndWhole(t *testing.T) {
	spec, ok := Lookup(tidebinderOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

// Countering an ability of a creature strips that creature's abilities for
// as long as the Tidebinder stays; when it leaves they come back.
func TestTidebinderCountersAnAbilityAndStripsItsCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	angel := pushCatalogPermanent(g, opp.ID, "Serra Angel", "Creature — Angel", serraAngelOracle, false)
	if !effectiveAbilitiesContain(t, g, angel, "flying") {
		t.Fatal("setup: Serra Angel should fly")
	}
	fired := false
	ability := taAnnounce(t, g, opp.ID, angel, "Serra Angel — do a thing", nil, &fired)

	binder := castAndResolveCreature(t, g, "Tishana's Tidebinder", "Creature — Merfolk Wizard", tidebinderOracle)
	answerPickTarget(t, g, ability)
	passPriorityAroundTable(t, g)

	if fired {
		t.Error("the countered ability resolved")
	}
	if effectiveAbilitiesContain(t, g, angel, "flying") {
		t.Error("the creature kept its abilities")
	}
	if !g.Battlefield.Contains(binder) {
		t.Fatal("the Tidebinder is not on the battlefield")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(binder) })
	if !effectiveAbilitiesContain(t, g, angel, "flying") {
		t.Error("the abilities did not return when the Tidebinder left")
	}
	_ = me
}

// A land's ability is countered, and the land is left alone: only an
// artifact, creature or planeswalker loses its abilities.
func TestTidebinderLeavesALandsAbilitiesAlone(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := g.Seats[1]
	land := pushCatalogPermanent(g, opp.ID, "Their Land", "Land", "", false)
	fired := false
	ability := taAnnounce(t, g, opp.ID, land, "Their Land — do a thing", nil, &fired)

	castAndResolveCreature(t, g, "Tishana's Tidebinder", "Creature — Merfolk Wizard", tidebinderOracle)
	answerPickTarget(t, g, ability)
	passPriorityAroundTable(t, g)
	if fired {
		t.Error("the countered ability resolved")
	}
	var stripped bool
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == land {
				stripped = g.Battlefield.Cards[i].Effective().AbilitiesRemoved
			}
		}
	})
	if stripped {
		t.Error("a land lost its abilities")
	}
}

// "Up to one": with nothing on the stack the creature simply enters.
func TestTidebinderWithNothingToCounterJustEnters(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	binder := castAndResolveCreature(t, g, "Tishana's Tidebinder", "Creature — Merfolk Wizard", tidebinderOracle)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(binder) {
		t.Error("the Tidebinder did not enter")
	}
	_ = uuid.Nil
	_ = game.ZoneBattlefield
}

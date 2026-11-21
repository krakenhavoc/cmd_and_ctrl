package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// raffine_a_tidebinder_test.go — Tishana's Tidebinder (#3212, slice raffine-a).

const tidebinderOracle = "2993dc7d-723d-4a9b-94bd-4bb02a9f7243"

func TestTidebinderIsRegisteredAndWhole(t *testing.T) {
	spec, ok := effects.Lookup(tidebinderOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

// Countering an ability of a creature strips that creature's abilities for
// as long as the Tidebinder stays; when it leaves they come back.
func TestTidebinderCountersAnAbilityAndStripsItsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	angel := effectstest.PushCatalogPermanent(g, opp.ID, "Serra Angel", "Creature — Angel", effectstest.SerraAngelOracle, false)
	if !effectstest.EffectiveAbilitiesContain(t, g, angel, "flying") {
		t.Fatal("setup: Serra Angel should fly")
	}
	fired := false
	ability := effectstest.TaAnnounce(t, g, opp.ID, angel, "Serra Angel — do a thing", nil, &fired)

	binder := effectstest.CastAndResolveCreature(t, g, "Tishana's Tidebinder", "Creature — Merfolk Wizard", tidebinderOracle)
	effectstest.AnswerPickTarget(t, g, ability)
	effectstest.PassPriorityAroundTable(t, g)

	if fired {
		t.Error("the countered ability resolved")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, angel, "flying") {
		t.Error("the creature kept its abilities")
	}
	if !g.Battlefield.Contains(binder) {
		t.Fatal("the Tidebinder is not on the battlefield")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(binder) })
	if !effectstest.EffectiveAbilitiesContain(t, g, angel, "flying") {
		t.Error("the abilities did not return when the Tidebinder left")
	}
	_ = me
}

// A land's ability is countered, and the land is left alone: only an
// artifact, creature or planeswalker loses its abilities.
func TestTidebinderLeavesALandsAbilitiesAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := g.Seats[1]
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Their Land", "Land", "", false)
	fired := false
	ability := effectstest.TaAnnounce(t, g, opp.ID, land, "Their Land — do a thing", nil, &fired)

	effectstest.CastAndResolveCreature(t, g, "Tishana's Tidebinder", "Creature — Merfolk Wizard", tidebinderOracle)
	effectstest.AnswerPickTarget(t, g, ability)
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	binder := effectstest.CastAndResolveCreature(t, g, "Tishana's Tidebinder", "Creature — Merfolk Wizard", tidebinderOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(binder) {
		t.Error("the Tidebinder did not enter")
	}
	_ = uuid.Nil
	_ = game.ZoneBattlefield
}

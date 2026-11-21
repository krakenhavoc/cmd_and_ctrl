package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestUmbralCollarZealotSacrificesAnotherPermanentToSurveil checks the
// sacrifice-outlet cost and that it can pay with a creature or an
// artifact.
func TestUmbralCollarZealotSacrificesAnotherPermanentToSurveil(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Bottom Me")
	zealot := effectstest.PushCatalogPermanent(g, me.ID, "Umbral Collar Zealot", "Creature — Human Cleric",
		effectstest.UmbralCollarZealotOracle, false)
	fodder := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Artifact",
		Owner: me.ID, Controller: me.ID,
	})

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, zealot, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed artifact should be gone")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SurveilChoiceFor(g, me.ID); c == nil {
		t.Fatalf("no surveil-1 prompt: %+v", g.PendingChoices)
	}
}

// TestUmbralCollarZealotCannotSacrificeItself — "another".
func TestUmbralCollarZealotCannotSacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	zealot := effectstest.PushCatalogPermanent(g, me.ID, "Umbral Collar Zealot", "Creature — Human Cleric",
		effectstest.UmbralCollarZealotOracle, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	err := g.ActivateCatalogAbility(me.ID, zealot, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{zealot},
	})
	if err == nil {
		t.Error("Umbral Collar Zealot should not be a legal sacrifice for its own ability")
	}
}

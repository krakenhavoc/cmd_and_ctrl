package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const umbralCollarZealotOracle = "12db6263-75c2-442f-a1a5-7af7915f8f9f"

// TestUmbralCollarZealotSacrificesAnotherPermanentToSurveil checks the
// sacrifice-outlet cost and that it can pay with a creature or an
// artifact.
func TestUmbralCollarZealotSacrificesAnotherPermanentToSurveil(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Bottom Me")
	zealot := pushCatalogPermanent(g, me.ID, "Umbral Collar Zealot", "Creature — Human Cleric",
		umbralCollarZealotOracle, false)
	fodder := pushBattlefieldCardWithTimestamp(g, game.Card{
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
	passPriorityAroundTable(t, g)
	if c := surveilChoiceFor(g, me.ID); c == nil {
		t.Fatalf("no surveil-1 prompt: %+v", g.PendingChoices)
	}
}

// TestUmbralCollarZealotCannotSacrificeItself — "another".
func TestUmbralCollarZealotCannotSacrificeItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	zealot := pushCatalogPermanent(g, me.ID, "Umbral Collar Zealot", "Creature — Human Cleric",
		umbralCollarZealotOracle, false)

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

package testsm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const massMutinyOracle = "d96763a0-6a6e-4520-899a-468b4bb307c8"

func TestMassMutinyIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(massMutinyOracle)
	if !ok || spec.Name != "Mass Mutiny" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func TestMassMutinyStealsUntapsAndHastesOneCreaturePerOpponentUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a, b := g.Seats[1], g.Seats[2]
	stolenA := effectstest.TdCreature(g, a.ID, "Stolen A", 2, 2)
	effectstest.TdCreature(g, a.ID, "Left A", 1, 1)
	stolenB := effectstest.TdCreature(g, b.ID, "Stolen B", 3, 3)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(stolenA) })

	effectstest.TdCastFromHand(t, g, game.Card{Name: "Mass Mutiny", TypeLine: "Sorcery", OracleID: massMutinyOracle, ManaCost: "{3}{R}{R}"},
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: stolenA}, {Kind: game.TargetCard, ID: stolenB}}})
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{stolenA, stolenB} {
		c, _ := g.LookupCardForEffect(id)
		if c.Controller != me.ID {
			t.Errorf("%s controller = %v, want the caster", c.Name, c.Controller)
		}
		if c.Tapped {
			t.Errorf("%s is still tapped", c.Name)
		}
		if !game.HasKeyword(&c, "haste") {
			t.Errorf("%s lacks haste", c.Name)
		}
	}
	// Only the picked creatures moved.
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Left A" && c.Controller != a.ID {
			t.Error("the unpicked creature changed hands")
		}
	}
	// It all ends with the turn: control reverts at cleanup.
	effectstest.EndTurn(t, g)
	for _, id := range []uuid.UUID{stolenA, stolenB} {
		c, _ := g.LookupCardForEffect(id)
		if c.Controller == me.ID {
			t.Errorf("%s was not returned at end of turn", c.Name)
		}
	}
}

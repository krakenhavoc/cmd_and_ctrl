package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// baActivate activates ability 0 and resolves it.
func BaActivate(t *testing.T, g *game.Game, me *game.Player, card uuid.UUID, targets []game.TargetRef) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	PassPriorityAroundTable(t, g)
}

// baSkipCombat removes the turn's combat phase from the plan still to
// come, as a skipped combat phase leaves it: the cursor goes from the
// precombat main phase to the postcombat main phase and a declare
// attackers step never begins (#4180).
func BaSkipCombat(t *testing.T, g *game.Game) {
	t.Helper()
	g.WithWriteLock(func() {
		var plan []game.PlannedStep
		for _, ps := range g.TurnPlan {
			if game.PhaseOf(ps.Step) != game.PhaseCombat {
				plan = append(plan, ps)
			}
		}
		if len(plan) == len(g.TurnPlan) {
			t.Fatal("setup: the plan has no combat phase to skip")
		}
		g.TurnPlan = plan
	})
}

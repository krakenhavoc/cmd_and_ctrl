package testsav

import (
	"slices"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blinding Angel (#5838): "Whenever this creature deals combat damage to
// a player, that player skips their next combat phase."

const blindingAngelOracle = "cb479ce1-59e8-4531-9a21-f7c97d334817"

func TestBlindingAngelIsRegisteredFull(t *testing.T) {
	def, ok := effects.Lookup(blindingAngelOracle)
	if !ok {
		t.Fatal("Blinding Angel is not registered")
	}
	if def.Name != "Blinding Angel" || def.Completeness != effects.CompletenessFull {
		t.Errorf("name %q, completeness %v", def.Name, def.Completeness)
	}
}

// blindingAngelTurnSteps walks to `seat`'s next upkeep and through that
// whole turn, and returns the steps that began in it.
func blindingAngelTurnSteps(t *testing.T, g *game.Game, seat int) []game.Step {
	t.Helper()
	effectstest.AdvanceToStepOf(t, g, seat, game.StepUpkeep)
	var since uint64
	for _, ev := range g.Events {
		since = max(since, ev.Seq)
	}
	for i := 0; i < 64 && g.Turn.ActiveSeat == seat; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	var out []game.Step
	for _, ev := range g.Events {
		if ev.Seq <= since || ev.Kind != game.EventStepBegan {
			continue
		}
		out = append(out, ev.Step)
		if ev.Step == game.StepCleanup {
			break
		}
	}
	return out
}

func blindingAngelHasCombat(steps []game.Step) bool {
	for _, s := range steps {
		if game.PhaseOf(s) == game.PhaseCombat {
			return true
		}
	}
	return false
}

// The Angel connects in a real combat: the damaged player's next turn has
// no combat phase but keeps its main phases, the attacker owes nothing,
// and the skip is used up, so the turn after has combat again.
func TestBlindingAngelSkipsTheDamagedPlayersNextCombatPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := effectstest.B12Push(g, me.ID, "Blinding Angel", "Creature — Angel", blindingAngelOracle, 2, 4)
	if !effectstest.HasEffectiveKeyword(t, g, angel, "flying") {
		t.Fatal("Blinding Angel has no flying")
	}
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, angel)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Fatalf("defender at %d life, want %d", opp.Life, life-2)
	}
	if n := g.NextCombatPhaseSkipsForLocked(opp.ID); n != 1 {
		t.Fatalf("the damaged player owes %d combat skips, want 1", n)
	}
	if n := g.NextCombatPhaseSkipsForLocked(me.ID); n != 0 {
		t.Errorf("the attacker owes %d combat skips, want 0", n)
	}
	steps := blindingAngelTurnSteps(t, g, 1)
	if blindingAngelHasCombat(steps) {
		t.Fatalf("the damaged player's next turn had combat: %v", steps)
	}
	for _, want := range []game.Step{game.StepPrecombatMain, game.StepPostcombatMain, game.StepEnd} {
		if !slices.Contains(steps, want) {
			t.Errorf("the damaged player's turn lost its %s: %v", want, steps)
		}
	}
	if n := g.NextCombatPhaseSkipsForLocked(opp.ID); n != 0 {
		t.Errorf("%d skips left after the skipped combat, want 0", n)
	}
	if steps := blindingAngelTurnSteps(t, g, 1); !blindingAngelHasCombat(steps) {
		t.Errorf("the damaged player's turn after that had no combat: %v", steps)
	}
}

// Noncombat damage does not trigger it.
func TestBlindingAngelIgnoresNoncombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	angel := effectstest.B12Push(g, me.ID, "Blinding Angel", "Creature — Angel", blindingAngelOracle, 2, 4)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: angel, Target: opp.ID, Amount: 2})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := g.NextCombatPhaseSkipsForLocked(opp.ID); n != 0 {
		t.Errorf("noncombat damage made the player owe %d combat skips", n)
	}
}

package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Savage Stomp — #4380, S76. Uses the existing catalog seams.
func init() {
	effects.Register(effects.Spec{
		OracleID: "2ecda58d-9aa9-4857-bdc5-b6b7b70a79c0", Name: "Savage Stomp", Completeness: effects.CompletenessFull,
		Targets:           effects.Clauses(effects.TargetCreature("target creature you control", effects.YouControl()), effects.TargetCreature("target creature you don't control", effects.Not(effects.YouControl()))),
		SelfCostModifiers: []game.CostModifier{effects.CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a Dinosaur you control.", effects.And(effects.Creature(), effects.YouControl(), effects.HasSubtype("Dinosaur")))},
		OnResolve:         effects.ResolveSteps(effects.Instruction(effects.CounterOnFirstCreatureTarget), effects.Instruction(effects.FightTargetClauses)),
	})
}

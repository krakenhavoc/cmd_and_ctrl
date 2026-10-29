package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Savage Stomp — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "2ecda58d-9aa9-4857-bdc5-b6b7b70a79c0", Name: "Savage Stomp", Completeness: CompletenessFull,
		Targets:           Clauses(TargetCreature("target creature you control", YouControl()), TargetCreature("target creature you don't control", Not(YouControl()))),
		SelfCostModifiers: []game.CostModifier{CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a Dinosaur you control.", And(Creature(), YouControl(), HasSubtype("Dinosaur")))},
		OnResolve:         ResolveSteps(Instruction(counterOnFirstCreatureTarget), Instruction(fightTargetClauses)),
	})
}

package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hunt the Weak — Sorcery {3}{G} (Magic 2014, #5563):
//
//	"Put a +1/+1 counter on target creature you control. Then that creature fights target creature you don't control. (Each deals damage equal to its power to the other.)"
//
// Kapow!'s counter-then-fight: the counter lands first, so the fight uses the
// bigger power (CR 701.14).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37864d13-2aaa-477b-8453-a42f77ab4437",
		Name:         "Hunt the Weak",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls())),
		OnResolve: effects.ResolveSteps(effects.Instruction(effects.CounterOnFirstCreatureTarget), effects.Instruction(effects.FightTargetClauses)),
	})
}

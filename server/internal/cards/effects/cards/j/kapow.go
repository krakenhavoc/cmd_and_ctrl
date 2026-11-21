package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kapow! — Sorcery {2}{G} (Marvel's Spider-Man, #4460):
//
//	"Put a +1/+1 counter on target creature you control. It fights target
//	 creature an opponent controls."
//
// Savage Stomp's two steps without its Dinosaur discount: the counter lands
// first, so the fight uses the grown power, and the fight is skipped if
// either target is no longer legal (CR 608.2b) while the counter still
// lands on a legal creature of yours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6eabbff-cc0c-4000-8feb-4826e1b6bc66",
		Name:         "Kapow!",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		OnResolve: effects.ResolveSteps(effects.Instruction(effects.CounterOnFirstCreatureTarget), effects.Instruction(effects.FightTargetClauses)),
	})
}

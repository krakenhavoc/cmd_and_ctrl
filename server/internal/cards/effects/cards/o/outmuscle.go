package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Outmuscle — Sorcery {3}{G} (Throne of Eldraine, #5516):
//
//	"Put a +1/+1 counter on target creature you control, then it fights
//	 target creature you don't control.
//	 Adamant — If at least three green mana was spent to cast this spell, the
//	 creature you control gains indestructible until end of turn."
//
// Hunt the Weak's counter-then-fight with the adamant rider as a third step,
// after the fight, in printed order: the indestructibility does not protect
// the creature in the fight itself. It goes on the creature you control only
// while it is still legal as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "185d84f4-6095-46ef-917f-3d662a86aa81",
		Name:           "Outmuscle",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"G"},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls())),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.CounterOnFirstCreatureTarget),
			effects.Instruction(effects.FightTargetClauses),
			effects.EldB1IfAdamantSpent("G", effects.OldB140IndestructibleToYourTarget{}),
		),
	})
}

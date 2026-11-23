package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Reluctant Role Model — Creature — Human Survivor {1}{W}, 2/2 (Duskmourn:
// House of Horror, #4466):
//
//	"Survival — At the beginning of your second main phase, if this
//	 creature is tapped, put a flying, lifelink, or +1/+1 counter on it.
//	 Whenever this creature or another creature you control dies, if it had
//	 counters on it, put those counters on up to one target creature."
//
// Survival is survival.go's trigger, the "if" an intervening if asked as the
// phase begins and again as it resolves; the counter is the controller's
// pick of the three as it resolves (Aragorn, Company Leader's picker).
//
// The death trigger is one registered trigger for the creature itself and
// every other creature you control (the source included, CR 603.10a). Its
// condition reads the counters the creature last had (the intervening if,
// CR 603.4), and the ability puts the same number of each kind on up to one
// target creature as it resolves, so a replacement that changes counters
// applies, -1/-1 counters come along, and two Role Models' triggers are
// separate. A Role Model that died together with the others still triggers
// for each of them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8d24c12a-a300-4199-b2a9-3d5ead8df617",
		Name:         "Reluctant Role Model",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Survival("Reluctant Role Model — put a flying, lifelink, or +1/+1 counter on it", effects.StdB26RoleModelPick),
			effects.Targeting(
				effects.On(game.EventLTB, effects.StdB26RoleModelCreatureDied,
					"Reluctant Role Model — put those counters on up to one target creature", effects.StdB26RoleModelMoveCounters),
				effects.TargetCreature("up to one target creature").WithCount(0, 1)),
		},
	})
}

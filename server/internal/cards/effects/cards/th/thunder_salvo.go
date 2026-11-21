package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thunder Salvo — Instant {1}{R} (Outlaws of Thunder Junction, #4468):
//
//	"Thunder Salvo deals X damage to target creature, where X is 2 plus
//	 the number of other spells you've cast this turn."
//
// X is read as the spell resolves, from its controller's cast tally
// (CastTally.Total), less one for this spell's own cast. A copy was never
// cast (CR 707.10), so it subtracts nothing and counts the original among
// the "other spells", as Call Forth the Tempest's tally does. A target
// that is gone or no longer a creature takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3ce7c83-986f-410a-ae00-6c1c1f918a2b",
		Name:         "Thunder Salvo",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 2)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			other := ctx.Game.CastTallyFor(ctx.Controller()).Total
			if !item.IsCopy {
				other--
			}
			return effects.DamageToFirstTargetStep(2+max(other, 0))(item, ctx)
		},
	})
}

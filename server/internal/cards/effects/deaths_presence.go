package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Death's Presence — Enchantment {5}{G}:
//
//	"Whenever a creature you control dies, put X +1/+1 counters on
//	 target creature you control, where X is the power of the creature
//	 that died."
//
// X is the dead creature's last-known power (CR 608.2h), read at
// resolution off the trigger's permanent record. The target is chosen
// when the trigger goes on the stack; with no creature left to target the
// trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d0db5996-f749-48d2-8fb3-a69104113828",
		Name:         "Death's Presence",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverACreatureYouControlDies("Death's Presence — put X +1/+1 counters on target creature you control",
					Do(OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
						info, ok := ctx.TriggeringPermanent()
						if !ok || info.Power <= 0 {
							return nil
						}
						return AddCounter{Target: id, Kind: game.CounterPlusOne, N: info.Power}.Apply(ctx)
					}))),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

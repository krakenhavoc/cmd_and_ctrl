package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aetherflux Reservoir — Artifact {4} (EDHREC rank 368):
//
//	"Whenever you cast a spell, you gain 1 life for each spell you've
//	 cast this turn.
//	 Pay 50 life: This artifact deals 50 damage to any target."
//
// The trigger is a cast trigger (YouCast(nil)). The count is read as
// it resolves (CR 608.2h), off the per-turn cast tally, which the cast
// path bumps before EventCast: the spell that triggered it is counted,
// so the first spell of a turn gains 1, the second 2, and a spell cast
// in response to the trigger is counted as well, as the printed "you've
// cast this turn" asks. A copy is not cast and neither triggers nor
// counts (CR 707.10).
//
// "Pay 50 life" is a cost (CR 119.4, CR 118.3): a player with less than
// 50 life can't activate it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f15125e-ae89-471e-b055-226a783bc7c1",
		Name:         "Aetherflux Reservoir",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCast(nil),
				"Aetherflux Reservoir — gain 1 life for each spell you've cast this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.GainLife{Amount: g.CastTallyFor(ctx.Controller()).Total}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 50 life: This artifact deals 50 damage to any target.",
			Cost:    effects.PayLife(50),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 50)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(50),
		}},
	})
}

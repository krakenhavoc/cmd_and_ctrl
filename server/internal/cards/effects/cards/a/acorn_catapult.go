package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Acorn Catapult — Artifact {4} (Commander 2011, #5574):
//
//	"{1}, {T}: This artifact deals 1 damage to any target. That permanent's
//	 controller or that player creates a 1/1 green Squirrel creature token."
//
// The controller of a permanent target is read before the damage, so a
// creature the ping kills still gives its last controller the Squirrel, and
// the Squirrel is made even when the damage was prevented. A target that
// became illegal means the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56d13879-2dc5-4d84-918b-4bd1a5fc2a33",
		Name:         "Acorn Catapult",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: This artifact deals 1 damage to any target. That permanent's controller or that player creates a 1/1 green Squirrel creature token.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				legal := ctx.LegalTargets()
				if len(legal) == 0 {
					return nil
				}
				t := legal[0]
				receiver := t.ID
				if t.Kind != game.TargetPlayer {
					controller, ok := effects.ControllerOfTarget(ctx, t.ID)
					if !ok {
						return nil
					}
					receiver = controller
				}
				if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.CreateToken{Controller: receiver, Template: effects.TokenCard("1/1 green Squirrel"), N: 1}.Apply(ctx)
			},
		}},
	})
}

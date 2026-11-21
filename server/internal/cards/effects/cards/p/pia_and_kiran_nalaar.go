package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pia and Kiran Nalaar — Legendary Creature — Human Artificer {2}{R}{R},
// 2/2:
//
//	"When Pia and Kiran Nalaar enters, create two 1/1 colorless Thopter
//	 artifact creature tokens with flying.
//	 {2}{R}, Sacrifice an artifact: Pia and Kiran Nalaar deals 2 damage to
//	 any target."
//
// Whirler Rogue's two Thopters and Goblin Bombardment's sacrifice-for-damage
// shape: the artifact is a cost paid on announce (it may be a Thopter, or
// the one the damage would otherwise be waiting on), and the damage comes
// from the creature, last-known information if it has left (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86084ed9-b8bb-4289-9579-6056194787bc",
		Name:         "Pia and Kiran Nalaar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pia and Kiran Nalaar — create two 1/1 colorless Thopter artifact creature tokens with flying",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: 2})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}, Sacrifice an artifact: Pia and Kiran Nalaar deals 2 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.DealDamage{
					Source: ctx.Source(),
					Target: item.Targets[0].ID,
					Amount: 2,
				}.Apply(ctx)
			},
		}},
	})
}

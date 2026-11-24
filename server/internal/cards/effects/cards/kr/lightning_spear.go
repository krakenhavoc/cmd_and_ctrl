package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Spear — Artifact — Equipment {1}{R} (Modern Horizons 2, #5501):
//
//	"Equipped creature gets +1/+0 and has trample.
//	 {2}{R}, Sacrifice this Equipment: It deals 3 damage to any target.
//	 Equip {1}"
//
// The Spear deals the damage as it last existed (CR 608.2h), so a Spear
// sacrificed as the cost still deals it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2202faf1-5638-46bf-8bf4-91870b669717",
		Name:         "Lightning Spear",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("trample"),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}{R}, Sacrifice this Equipment: It deals 3 damage to any target.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
				Cost:    effects.Plus(effects.ManaCost("{2}{R}"), effects.SacrificeThis()),
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(3),
			},
			effects.EquipAbility("{1}"),
		},
	})
}

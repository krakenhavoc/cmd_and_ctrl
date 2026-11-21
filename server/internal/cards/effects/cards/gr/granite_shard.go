package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Granite Shard — Artifact {3} (Mirrodin, #5606):
//
//	"{3}, {T} or {R}, {T}: This artifact deals 1 damage to any target."
//
// Crystal Shard's shape: one ability with a choice of cost, written as
// two rows so the activator picks the cost by picking the row. Both rows
// resolve as Anaba Shaman's ping: 1 damage from the Shard to the target,
// which is checked again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	const effect = "This artifact deals 1 damage to any target."
	effects.Register(effects.Spec{
		OracleID:     "db81c783-a622-41e1-9dd6-47fab1d76967",
		Name:         "Granite Shard",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}, {T}: " + effect,
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(1),
			},
			{
				Label:   "{R}, {T}: " + effect,
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 1)),
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(1),
			},
		},
	})
}

package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pyrite Spellbomb — Artifact {1} (Mirrodin, #5606):
//
//	"{R}, Sacrifice this artifact: It deals 2 damage to any target.
//	 {1}, Sacrifice this artifact: Draw a card."
//
// The first row is Aeolipile's: 2 damage from the Spellbomb, read as
// it last existed (CR 608.2h), to a target checked again as the
// ability resolves (CR 608.2b).
// The second row is Aether Spellbomb's: {1}, sacrifice it, draw a card
// (mrdB5SpellbombDraw). The sacrifice is part of the cost, so the
// Spellbomb is gone before either row resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c10cae2-951a-4f4f-94e4-8713b58d07dd",
		Name:         "Pyrite Spellbomb",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}, Sacrifice this artifact: It deals 2 damage to any target.",
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Cost:    effects.Plus(effects.ManaCost("{R}"), effects.SacrificeThis()),
				Targets: effects.TargetAny(),
				Effect:  effects.DealDamageToTheTarget(2),
			},
			effects.MrdB5SpellbombDraw(),
		},
	})
}

package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Acidic Sliver — Creature — Sliver {B}{R}, 2/2 (Stronghold, S82 OLD batch 747,
// #5627):
//
//	"All Slivers have "{2}, Sacrifice this permanent: This permanent deals 2
//	 damage to any target.""
//
// Firewake Sliver's grant (ADR 0093): an activated ability given to every
// Sliver at the table, whose cost sacrifices the Sliver that has it. Only that
// Sliver's controller can activate it (CR 602.2), it is the damage source, and
// the target is any target.
//
// No simplification.
func init() {
	const grant = "acidic-sliver/sacrifice-deal-2"
	effects.Register(effects.Spec{
		OracleID:     "a149eeb1-007c-435d-9d55-34af205090e0",
		Name:         "Acidic Sliver",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{2}, Sacrifice this permanent: This permanent deals 2 damage to any target",
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeThis()),
				Targets: effects.TargetAny(),
				Purpose: effects.ForTargets(effects.DamageToTarget(0, 2)),
				Effect:  effects.DealDamageToTheTarget(2),
			}},
			Text: "{2}, Sacrifice this permanent: This permanent deals 2 damage to any target.",
		}},
		Static: []game.StaticAbility{
			effects.TribalAbilityGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, grant),
		},
	})
}

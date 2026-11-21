package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// High Noon — Enchantment {1}{W} (Outlaws of Thunder Junction, #4468):
//
//	"Each player can't cast more than one spell each turn.
//	 {4}{R}, Sacrifice this enchantment: It deals 5 damage to any target."
//
// The cast limit is Rule of Law's (EachPlayerMaxSpellsPerTurn). The damage
// is dealt by the sacrificed enchantment, read as it last existed
// (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c8cbf0f-0288-4b5e-bd4f-0f208670b6b5",
		Name:         "High Noon",
		Completeness: effects.CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			effects.EachPlayerMaxSpellsPerTurn(1,
				"High Noon — each player can't cast more than one spell each turn."),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{R}, Sacrifice this enchantment: It deals 5 damage to any target.",
			Cost:    effects.Plus(effects.ManaCost("{4}{R}"), effects.SacrificeThis()),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 5)),
			Targets: effects.TargetAny(),
			Effect:  effects.SourceDealsFiveToTheChosenTarget,
		}},
	})
}

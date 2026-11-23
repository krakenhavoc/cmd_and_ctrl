package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vassal's Duty — Enchantment {3}{W}:
//
//	"{1}: The next 1 damage that would be dealt to target legendary
//	 creature you control this turn is dealt to you instead."
//
// ADR 0108 §9 (#1905): a 1-point charged redirection (CR 615.7) from the
// target to the ability's controller.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7d84e667-2f14-437c-be4f-161b98d59341",
		Name:         "Vassal's Duty",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.RedirectRow("{1}: The next 1 damage that would be dealt to target legendary creature you control this turn is dealt to you instead.",
				effects.ManaCost("{1}"), effects.TargetCreature("target legendary creature you control", effects.Legendary(), effects.YouControl()),
				effects.RedirectDamage{Protect: effects.ShieldTheTarget, Amount: 1, To: effects.RedirectToYou}),
		},
	})
}

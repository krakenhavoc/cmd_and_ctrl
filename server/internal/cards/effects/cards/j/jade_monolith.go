package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jade Monolith — Artifact {4}:
//
//	"{1}: The next time a source of your choice would deal damage to
//	 target creature this turn, that source deals that damage to you
//	 instead."
//
// ADR 0108 §9 (#1905): a "next time" redirection from a source chosen as
// the ability resolves (CR 609.7a), from the target to the ability's
// controller.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1e105ab7-fb10-4cfd-ac2f-5e11488cf1b0",
		Name:         "Jade Monolith",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.RedirectRow("{1}: The next time a source of your choice would deal damage to target creature this turn, that source deals that damage to you instead.",
				effects.ManaCost("{1}"), effects.TargetCreature("target creature"),
				effects.RedirectDamage{Choose: true, Protect: effects.ShieldTheTarget, Next: true, To: effects.RedirectToYou}),
		},
	})
}

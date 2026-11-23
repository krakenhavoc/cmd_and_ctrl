package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Beacon of Destiny — Creature — Human Cleric {1}{W}, 1/3:
//
//	"{T}: The next time a source of your choice would deal damage to you
//	 this turn, that damage is dealt to this creature instead."
//
// ADR 0108 §9 (#1905): a "next time" redirection from a source chosen as
// the ability resolves (CR 609.7a) to the Beacon as it is now. A Beacon
// that has left the battlefield by then redirects nothing (CR 614.9).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1ed2274-7774-4c65-a95b-28ae5e225994",
		Name:         "Beacon of Destiny",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.RedirectRow("{T}: The next time a source of your choice would deal damage to you this turn, that damage is dealt to this creature instead.",
				effects.TapCost(), nil, effects.RedirectDamage{Choose: true, Protect: effects.ShieldYou, Next: true, To: effects.RedirectToThis}),
		},
	})
}

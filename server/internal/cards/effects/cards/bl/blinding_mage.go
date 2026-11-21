package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blinding Mage — Creature — Human Wizard {1}{W}, 1/2 (Magic 2010, #5585):
//
//	"{W}, {T}: Tap target creature."
//
// Nebelgast Beguiler's tap ability: the target is chosen on activation and
// tapped as it resolves, if it is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "53a1faa1-ca67-4027-bee8-0f7c4584d62b",
		Name:         "Blinding Mage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}

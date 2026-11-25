package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Proud Mentor — Creature — Human Warrior {2}{W}, 1/1 (Battlebond, #5529):
//
//	"Partner with Impetuous Protege (When this creature enters, target player may put Impetuous Protege into their hand from their library, then shuffle.)"
//	"{W}, {T}: Tap target creature."
//
// Blinding Mage's activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ad114d8-2123-4267-b30b-6f4e755bb0de",
		Name:         "Proud Mentor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Proud Mentor", "Impetuous Protege"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}, {T}: Tap target creature.",
			Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}

package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Earth Kingdom Protectors — Creature — Human Soldier Ally {W}, 1/1
// (Avatar: The Last Airbender, #4459):
//
//	"Vigilance
//	 Sacrifice this creature: Another target Ally you control gains
//	 indestructible until end of turn."
//
// "Another" is object identity (Another, ExcludeSource): the Protectors
// can never pick themselves, and are gone by the time the ability
// resolves anyway, since the sacrifice is the cost. The Ally is any
// permanent you control with the type, read post-layer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3bfb3913-808d-444b-97c3-a90af3bc585b",
		Name:            "Earth Kingdom Protectors",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Another target Ally you control gains indestructible until end of turn.",
			Cost:    effects.SacrificeThis(),
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible")),
			Targets: effects.Another(effects.TargetPermanent("another target Ally you control", effects.HasSubtype("Ally"), effects.YouControl())),
			Effect:  effects.FirstTargetGainsUntilEndOfTurn(0, "Earth Kingdom Protectors — indestructible until end of turn", "indestructible"),
		}},
	})
}

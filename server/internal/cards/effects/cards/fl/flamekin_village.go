package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flamekin Village — Land (Lorwyn Eclipsed Commander, #5368):
//
//	"As this land enters, you may reveal an Elemental card from your hand. If
//	 you don't, this land enters tapped.
//	 {T}: Add {R}.
//	 {R}, {T}: Target creature gains haste until end of turn."
//
// The reveal-land entry over Elemental cards and Hanweir Battlements'
// haste ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "34a1eb04-08f6-49d8-a1d1-b987a76bd8b1",
		Name:         "Flamekin Village",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersTappedUnlessYouRevealFromHand("Flamekin Village", "an Elemental card", effects.Cmdr2S12IsElementalCard),
		},
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{R}", Label: "Add {R}"},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}: Target creature gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("haste", "Flamekin Village — haste until end of turn"),
		}},
	})
}

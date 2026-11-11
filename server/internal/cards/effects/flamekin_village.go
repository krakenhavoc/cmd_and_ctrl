package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "34a1eb04-08f6-49d8-a1d1-b987a76bd8b1",
		Name:         "Flamekin Village",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersTappedUnlessYouRevealFromHand("Flamekin Village", "an Elemental card", cmdr2S12IsElementalCard),
		},
		ManaAbilities: []ManaAbility{
			{Cost: ManaAbilityCost{Tap: true}, Produced: "{R}", Label: "Add {R}"},
		},
		Activated: []ActivatedAbility{{
			Label:   "{R}, {T}: Target creature gains haste until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    Plus(ManaCost("{R}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  ebTargetCreatureGainsUntilEOT("haste", "Flamekin Village — haste until end of turn"),
		}},
	})
}

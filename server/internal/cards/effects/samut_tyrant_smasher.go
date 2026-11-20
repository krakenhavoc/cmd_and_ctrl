package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Samut, Tyrant Smasher — Legendary Planeswalker — Samut {2}{R/G}{R/G}, loyalty 5 (War of the Spark, #5520):
//
//	"Creatures you control have haste.
//	 −1: Target creature gets +2/+1 and gains haste until end of turn. Scry 1."
//
// Starting loyalty is printed data (ADR 0032). The static is Urabrask the Hidden's haste
// grant; the −1 is Samut's Sprint's text (warPlusTwoOneAndHaste, then
// scry 1).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80405d7a-f533-44bd-ac85-890476cf2b78",
		Name:         "Samut, Tyrant Smasher",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(b16CreaturesYouControl, "haste"),
		},
		Activated: []ActivatedAbility{{
			Label:   "−1: Target creature gets +2/+1 and gains haste until end of turn. Scry 1.",
			Purpose: game.Purpose{Scry: 1, Targets: game.ForTargets(BenefitUntilEOT(0, 2, 1, "haste"))},
			Cost:    LoyaltyCost(-1),
			Targets: TargetCreature("target creature"),
			Effect:  Do(warPlusTwoOneAndHaste("Samut, Tyrant Smasher — +2/+1 and haste"), Scry{N: 1}),
		}},
	})
}

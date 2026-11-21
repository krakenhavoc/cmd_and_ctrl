package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "80405d7a-f533-44bd-ac85-890476cf2b78",
		Name:         "Samut, Tyrant Smasher",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(effects.B16CreaturesYouControl, "haste"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "−1: Target creature gets +2/+1 and gains haste until end of turn. Scry 1.",
			Purpose: game.Purpose{Scry: 1, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 2, 1, "haste"))},
			Cost:    effects.LoyaltyCost(-1),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.Do(effects.WarPlusTwoOneAndHaste("Samut, Tyrant Smasher — +2/+1 and haste"), effects.Scry{N: 1}),
		}},
	})
}

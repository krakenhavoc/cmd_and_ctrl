package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nyxborn Behemoth — Enchantment Creature — Beast {10}{G}{G}, 10/10
// (Commander Masters):
//
//	"This spell costs {X} less to cast, where X is the total mana value of
//	 noncreature enchantments you control.
//	 Trample
//	 {1}{G}, Sacrifice another enchantment: This creature gains indestructible
//	 until end of turn."
//
// The discount is Ghoultree's shape (CR 601.2f), reducing generic mana only,
// with an amount summed over the noncreature enchantments on your side. The
// sacrifice is a cost, and "another" keeps the Behemoth itself out of it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "58e20ca7-5fe3-49ff-bee0-0388722e3cab",
		Name:            "Nyxborn Behemoth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(gapB103NoncreatureEnchantmentValue,
				"This spell costs {X} less to cast, where X is the total mana value of noncreature enchantments you control."),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}, Sacrifice another enchantment: This creature gains indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive | game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{1}{G}"), SacrificeAnotherN(1, "another enchantment", Enchantment())),
			Effect:  thisGainsKeywordUntilEndOfTurn("indestructible", "Nyxborn Behemoth — indestructible until end of turn"),
		}},
	})
}

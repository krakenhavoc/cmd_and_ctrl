package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Go-Shintai of Hidden Cruelty — Legendary Enchantment Creature — Shrine {3}{B}, 2/2
// (Kamigawa: Neon Dynasty, #5493):
//
//	"Deathtouch
//	 At the beginning of your end step, you may pay {1}. When you do, destroy target creature with toughness X or less, where X is the number of Shrines you control."
//
// The opener is MayPay, asked as the end-step trigger resolves. When you do, the reflexive trigger (CR 603.12) takes its target after the payment: a creature whose toughness is X or less as it goes on the stack and again as it resolves.
// The Shrines are counted as the effect resolves (CR 608.2h), the
// Go-Shintai itself among them while it is on the battlefield; "Shrine"
// is read through the text lens (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ff8cf669-a38a-44ae-baff-420f59492371",
		Name:            "Go-Shintai of Hidden Cruelty",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			neoGoShintaiEndStep("Go-Shintai of Hidden Cruelty — destroy target creature with toughness X or less",
				"Go-Shintai of Hidden Cruelty — pay {1}?",
				neoGoShintaiReflexive("hidden-cruelty", "destroy target creature with toughness X or less", neoHiddenCrueltyBody)),
		},
	})
}

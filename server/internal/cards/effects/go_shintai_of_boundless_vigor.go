package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Go-Shintai of Boundless Vigor — Legendary Enchantment Creature — Shrine {1}{G}, 1/1
// (Kamigawa: Neon Dynasty, #5493):
//
//	"Trample
//	 At the beginning of your end step, you may pay {1}. When you do, put a +1/+1 counter on target Shrine for each Shrine you control."
//
// The opener is MayPay, asked as the end-step trigger resolves. When you do, the reflexive trigger (CR 603.12) takes its target Shrine after the payment (the Go-Shintai may be it).
// The Shrines are counted as the effect resolves (CR 608.2h), the
// Go-Shintai itself among them while it is on the battlefield; "Shrine"
// is read through the text lens (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "183cfd35-c18a-4838-94e5-1ad66e6bd325",
		Name:            "Go-Shintai of Boundless Vigor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			neoGoShintaiEndStep("Go-Shintai of Boundless Vigor — put a +1/+1 counter on target Shrine for each Shrine you control",
				"Go-Shintai of Boundless Vigor — pay {1}?",
				neoGoShintaiReflexive("boundless-vigor", "put a +1/+1 counter on target Shrine for each Shrine you control", neoBoundlessVigorBody)),
		},
	})
}

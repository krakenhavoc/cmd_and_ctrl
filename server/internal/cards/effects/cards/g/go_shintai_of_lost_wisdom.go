package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Go-Shintai of Lost Wisdom — Legendary Enchantment Creature — Shrine {1}{U}, 0/4
// (Kamigawa: Neon Dynasty, #5493):
//
//	"Flying
//	 At the beginning of your end step, you may pay {1}. When you do, target player mills X cards, where X is the number of Shrines you control."
//
// The opener is MayPay, asked as the end-step trigger resolves. When you do, the reflexive trigger (CR 603.12) takes its target player after the payment.
// The Shrines are counted as the effect resolves (CR 608.2h), the
// Go-Shintai itself among them while it is on the battlefield; "Shrine"
// is read through the text lens (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f6a2092a-093e-4cd6-9eb0-d8ed32da896c",
		Name:            "Go-Shintai of Lost Wisdom",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.NeoGoShintaiEndStep("Go-Shintai of Lost Wisdom — target player mills X cards",
				"Go-Shintai of Lost Wisdom — pay {1}?",
				effects.NeoGoShintaiReflexive("lost-wisdom", "target player mills X cards", effects.NeoLostWisdomBody)),
		},
	})
}

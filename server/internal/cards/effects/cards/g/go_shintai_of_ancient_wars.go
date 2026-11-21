package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Go-Shintai of Ancient Wars — Legendary Enchantment Creature — Shrine {2}{R}, 2/2
// (Kamigawa: Neon Dynasty, #5493):
//
//	"First strike
//	 At the beginning of your end step, you may pay {1}. When you do, Go-Shintai of Ancient Wars deals X damage to target player or planeswalker, where X is the number of Shrines you control."
//
// The opener is MayPay, asked as the end-step trigger resolves. When you do, the reflexive trigger (CR 603.12) takes its target after the payment; the Go-Shintai is the damage source.
// The Shrines are counted as the effect resolves (CR 608.2h), the
// Go-Shintai itself among them while it is on the battlefield; "Shrine"
// is read through the text lens (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d145e4d4-a190-4c0b-8dfd-401c0111e995",
		Name:            "Go-Shintai of Ancient Wars",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.NeoGoShintaiEndStep("Go-Shintai of Ancient Wars — deals X damage to target player or planeswalker",
				"Go-Shintai of Ancient Wars — pay {1}?",
				effects.NeoGoShintaiReflexive("ancient-wars", "deals X damage to target player or planeswalker", effects.NeoAncientWarsBody)),
		},
	})
}

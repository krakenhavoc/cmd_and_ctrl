package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Go-Shintai of Shared Purpose — Legendary Enchantment Creature — Shrine
// {3}{W}, 1/3 (Kamigawa: Neon Dynasty, #5493):
//
//	"Vigilance
//	 At the beginning of your end step, you may pay {1}. If you do, create
//	 a 1/1 colorless Spirit creature token for each Shrine you control."
//
// Honden of Life's Web's count and token. "If you do" is the same
// resolution, not a reflexive trigger (nothing targets): MayPay asks as the
// end-step trigger resolves and the paid branch makes the Spirits, the
// Shrines counted then (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4cab6ea8-4a80-4b34-b758-7f2aa4d540e4",
		Name:            "Go-Shintai of Shared Purpose",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.NeoGoShintaiEndStep("Go-Shintai of Shared Purpose — create a 1/1 Spirit for each Shrine you control",
				"Go-Shintai of Shared Purpose — pay {1} to create a Spirit for each Shrine you control?",
				neoSharedPurposePaid),
		},
	})
}

var neoSharedPurposePaid = effects.PayUnlessBranches("pay-unless/neo-go-shintai-shared-purpose",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		n := effects.NeoShrinesYouControl(ctx.Game, ctx.Item)
		if n <= 0 {
			return nil
		}
		return effects.CreateToken{Template: effects.TokenCard("1/1 colorless Spirit"), N: n}.Apply(ctx)
	}, nil)

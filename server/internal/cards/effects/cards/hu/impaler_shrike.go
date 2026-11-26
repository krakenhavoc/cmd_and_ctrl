package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Impaler Shrike — Creature — Phyrexian Bird {2}{U}{U}, 3/1:
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, you may sacrifice
//	 it. If you do, draw three cards."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "db98ee99-d4e5-4623-8446-5cce5d11dca0",
		Name:            "Impaler Shrike",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverThisDealsCombatDamageToAPlayer(
				"Impaler Shrike — you may sacrifice it; if you do, draw three cards",
				effects.SacrificeThisThen("then/s82-old-b461-impaler-shrike-draw", func(ctx *effects.Context) error {
					return effects.DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
				})),
				"Impaler Shrike — sacrifice it to draw three cards?"),
		},
	})
}

package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elsha, Threefold Master — Legendary Creature — Djinn Monk {U}{R}{W},
// 1/1 (Tarkir: Dragonstorm Commander):
//
//	"Trample
//	 Prowess
//	 Whenever Elsha deals combat damage to a player, create that many
//	 1/1 white Monk creature tokens with prowess."
//
// Ant-Man's one-token-per-point-of-damage trigger with the Monk token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d8ccbd72-6429-43cb-a0a9-ab56e97a2d53",
		Name:            "Elsha, Threefold Master",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "prowess"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Elsha — create that many 1/1 white Monk tokens with prowess",
				effects.Do(effects.CreateTokensForEach{
					Template: effects.TokenCard("1/1 white Monk with prowess"),
					Name:     "the combat damage dealt",
					Count:    func(ctx *effects.Context) int { return ctx.Trigger().Event.Amount },
				})),
		},
	})
}

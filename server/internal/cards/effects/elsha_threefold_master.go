package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "d8ccbd72-6429-43cb-a0a9-ab56e97a2d53",
		Name:            "Elsha, Threefold Master",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "prowess"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Elsha — create that many 1/1 white Monk tokens with prowess",
				Do(CreateTokensForEach{
					Template: TokenCard("1/1 white Monk with prowess"),
					Name:     "the combat damage dealt",
					Count:    func(ctx *Context) int { return ctx.Trigger().Event.Amount },
				})),
		},
	})
}

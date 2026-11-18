package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sporeback Wolf — Creature — Wolf {1}{G}, 2/2 (Innistrad: Crimson Vow):
//
//	"During your turn, this creature gets +0/+2."
//
// Dapper Shieldmate's layer 7c static, re-read on every pass, on only
// while the controller is the active player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6fd5f773-e53e-4a7d-9093-dddd9b2e741f",
		Name:         "Sporeback Wolf",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "During your turn, this creature gets +0/+2.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			},
			Apply: pumpBy(0, 2),
		}},
	})
}

package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hookblade Veteran — Creature — Human Assassin {U}, 1/2 (Assassin's
// Creed, #5466):
//
//	"During your turn, this creature has flying."
//
// Shuttle Ace's grant: a layer 6 keyword on itself, on only while its
// controller is the active player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bad2385b-6981-46fa-a012-d3992c3bc118",
		Name:         "Hookblade Veteran",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "flying"),
		},
	})
}

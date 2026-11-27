package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sheltering Prayers — Enchantment {W} (Prophecy, S82 OLD batch 674, #5617):
//
//	"Basic lands each player controls have shroud as long as that player
//	 controls three or fewer lands. (They can't be the targets of spells or
//	 abilities.)"
//
// A layer 6 keyword grant (Fangorn's KeywordGrant shape) to every basic land,
// whoever controls it, while its own controller controls three or fewer lands
// of any kind: each player's count is their own, read live, so it switches off
// for one player when they control a fourth land and not for the others.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44075eea-ab9d-4252-b77c-597d19a10fc4",
		Name:         "Sheltering Prayers",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.Labelled(effects.KeywordGrant(func(target *game.Card, g *game.Game, _ *game.Card) bool {
			if !effects.IsBasicLand(*target) {
				return false
			}
			return effects.CountControlled(g, target.Controller, func(c game.Card) bool { return c.IsLand() }) <= 3
		}, "shroud"), "Basic lands each player controls have shroud as long as that player controls three or fewer lands.")},
	})
}

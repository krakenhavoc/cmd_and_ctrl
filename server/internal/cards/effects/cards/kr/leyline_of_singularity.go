package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Leyline of Singularity — Enchantment {2}{U}{U} (Guildpact, S82 OLD batch
// 573, #5599):
//
//	"If this card is in your opening hand, you may begin the game with it on
//	 the battlefield.
//	 All nonland permanents are legendary."
//
// The opening-hand clause is the Leylines' shared one (BeginTheGameOnTheBattlefield).
// The static is a layer 4 supertype grant (CR 205.4b, CR 613.1d) to every
// nonland permanent of every player, the Leyline included, so the legend rule
// (CR 704.5j) applies to them through the layered view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc9f159b-984d-4b9f-8904-b38b5cb79636",
		Name:         "Leyline of Singularity",
		Completeness: effects.CompletenessFull,
		OpeningHand:  effects.BeginTheGameOnTheBattlefield(),
		Static: []game.StaticAbility{{
			Label: "All nonland permanents are legendary.",
			Layer: game.Layer4Type,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return !target.IsLand()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, s := range c.Supertypes {
					if s == "Legendary" {
						return
					}
				}
				c.Supertypes = append(c.Supertypes, "Legendary")
			},
		}},
	})
}
